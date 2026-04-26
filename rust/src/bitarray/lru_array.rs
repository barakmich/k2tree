use crate::bitarray::BitArray;
use crate::error::K2TreeError;
use std::cell::UnsafeCell;
use std::collections::BTreeMap;

/// Default cache distance in bits. This value was optimized experimentally.
/// It's the distance in bits between cache hits. It's a tradeoff between
/// leaning on the POPCNT instruction between known offsets in the cache
/// and the overhead of maintaining the LRU. If the LRU gets cheaper to
/// maintain, this may get decreased. If POPCNT gets faster, this may increase.
pub const DEFAULT_LRU_CACHE_DISTANCE: usize = 512;

/// LruArray is a generic LRU-cached BitArray wrapper that uses static dispatch.
/// The generic parameter allows for zero-cost abstraction over any BitArray implementation:
/// - LruArray<SliceArray>
/// - LruArray<LruArray<SliceArray>> (nested caching if needed)
/// - Any future BitArray implementation
///
/// The cache stores precomputed popcount values at strategic offsets to avoid
/// repeated POPCNT operations on large ranges. Uses BTreeMap for O(log n) sorted
/// access and range queries, with a simple counter-based eviction policy.
///
/// Uses interior mutability via UnsafeCell to allow caching through immutable references,
/// which is necessary for the K2Tree iterator API.
pub struct LruArray<T: BitArray> {
    /// The inner BitArray implementation (owned, not boxed for zero-cost abstraction)
    bits: T,
    /// Cache state (BTreeMap keeps keys sorted for fast range queries)
    /// Wrapped in UnsafeCell for zero-overhead interior mutability
    cache: UnsafeCell<CacheState>,
    /// Maximum cache size
    size: usize,
    /// Minimum distance (in bits) to cache a new entry
    cache_distance: usize,
}

/// Cache state using BTreeMap for sorted keys
struct CacheState {
    /// Sorted map of offset -> popcount
    entries: BTreeMap<usize, usize>,
    /// Access times for LRU eviction
    access_times: BTreeMap<usize, usize>,
    /// Current time counter
    tick: usize,
}

impl<T: BitArray> LruArray<T> {
    /// Creates a new LruArray wrapping the given BitArray with the specified cache size.
    /// Uses the default cache distance of 512 bits.
    pub fn new(bits: T, size: usize) -> Self {
        Self::with_cache_distance(bits, size, DEFAULT_LRU_CACHE_DISTANCE)
    }

    /// Creates a new LruArray with a custom cache distance.
    /// This is useful for benchmarking and tuning the cache parameters.
    pub fn with_cache_distance(bits: T, size: usize, cache_distance: usize) -> Self {
        LruArray {
            bits,
            cache: UnsafeCell::new(CacheState {
                entries: BTreeMap::new(),
                access_times: BTreeMap::new(),
                tick: 0,
            }),
            size,
            cache_distance,
        }
    }

    /// Returns the number of cache entries currently stored.
    pub fn cache_size(&self) -> usize {
        unsafe { (*self.cache.get()).entries.len() }
    }

    /// Returns cache statistics for debugging.
    pub fn cache_stats(&self) -> (usize, usize) {
        unsafe {
            let cache = &*self.cache.get();
            (cache.entries.len(), cache.tick)
        }
    }

    /// Helper function to compute popcount from 0 to `to`.
    ///
    /// SAFETY: This method uses unsafe code to access the cache through UnsafeCell.
    /// This is sound because:
    /// 1. count() takes &self, so the method looks immutable from the outside
    /// 2. We never create multiple mutable references - we only ever access the cache once at a time
    /// 3. The cache is never accessed by external code, only through this struct's methods
    /// 4. Rust's borrow checker ensures no aliasing at the struct level
    fn zero_count(&self, to: usize) -> usize {
        unsafe {
            let cache = &mut *self.cache.get();

            // Check if we have an exact match in cache
            if let Some(&cached_count) = cache.entries.get(&to) {
                // Update access time
                cache.tick += 1;
                cache.access_times.insert(to, cache.tick);
                return cached_count;
            }

            // Find the closest cached offset using BTreeMap's range queries (O(log n))
            let (count, at) = self.find_closest_cached_btree(cache, to);

            // Calculate the result using the cached value
            let val = if at == to {
                count
            } else if at < to {
                count + self.bits.count(at, to)
            } else {
                count - self.bits.count(to, at)
            };

            // Cache if far enough away from the closest cached offset
            if (to as isize - at as isize).abs() as usize > self.cache_distance {
                // Evict LRU entry if cache is full
                if cache.entries.len() >= self.size {
                    if let Some((&lru_offset, _)) = cache.access_times.iter().next() {
                        cache.entries.remove(&lru_offset);
                        cache.access_times.remove(&lru_offset);
                    }
                }

                // Insert new entry
                cache.tick += 1;
                cache.entries.insert(to, val);
                cache.access_times.insert(to, cache.tick);
            }

            val
        }
    }

    /// Find the closest cached offset using BTreeMap range queries.
    /// Returns (count, offset) of the closest entry, or (0, 0) if cache is empty.
    ///
    /// SAFETY: Must be called with a valid mutable reference to the cache.
    unsafe fn find_closest_cached_btree(
        &self,
        cache: &mut CacheState,
        to: usize,
    ) -> (usize, usize) {
        if cache.entries.is_empty() {
            return (0, 0);
        }

        // Find entries <= to and entries > to
        let lower = cache.entries.range(..=to).next_back();
        let upper = cache.entries.range((to + 1)..).next();

        // Choose the closer one
        let (best_offset, best_count) = match (lower, upper) {
            (Some((&lo_off, &lo_count)), Some((&up_off, &up_count))) => {
                let lo_dist = to - lo_off;
                let up_dist = up_off - to;
                if lo_dist <= up_dist {
                    (lo_off, lo_count)
                } else {
                    (up_off, up_count)
                }
            }
            (Some((&lo_off, &lo_count)), None) => (lo_off, lo_count),
            (None, Some((&up_off, &up_count))) => (up_off, up_count),
            (None, None) => (0, 0),
        };

        // Update access time
        if best_offset > 0 {
            cache.tick += 1;
            cache.access_times.insert(best_offset, cache.tick);
        }

        (best_count, best_offset)
    }
}

impl<T: BitArray> BitArray for LruArray<T> {
    #[inline(always)]
    fn len(&self) -> usize {
        self.bits.len()
    }

    fn set(&mut self, at: usize, val: bool) {
        let cur = self.bits.get(at);
        if cur == val {
            return;
        }

        self.bits.set(at, val);

        // Update cached counts for all offsets > at
        unsafe {
            let cache = &mut *self.cache.get();
            let delta: isize = if val { 1 } else { -1 };

            // Collect all keys that need updating (offsets > at)
            let keys_to_update: Vec<usize> = cache
                .entries
                .range((at + 1)..)
                .map(|(&offset, _)| offset)
                .collect();

            // Update each affected cache entry
            for offset in keys_to_update {
                if let Some(&count) = cache.entries.get(&offset) {
                    let new_count = if delta > 0 {
                        count + 1
                    } else {
                        count.saturating_sub(1)
                    };
                    cache.entries.insert(offset, new_count);
                }
            }
        }
    }

    #[inline(always)]
    fn get(&self, at: usize) -> bool {
        self.bits.get(at)
    }

    fn count(&self, from: usize, to: usize) -> usize {
        if from == to {
            return 0;
        }

        let result = self.zero_count(to);
        if from != 0 {
            let subresult = self.zero_count(from);
            result - subresult
        } else {
            result
        }
    }

    #[inline(always)]
    fn total(&self) -> usize {
        self.bits.total()
    }

    fn insert(&mut self, n: usize, at: usize) -> Result<(), K2TreeError> {
        self.bits.insert(n, at)?;

        // Adjust cached offsets: all offsets >= at need to shift by n
        unsafe {
            let cache = &mut *self.cache.get();

            // Collect entries that need updating (offsets >= at)
            let entries_to_update: Vec<(usize, usize)> = cache
                .entries
                .range(at..)
                .map(|(&offset, &count)| (offset, count))
                .collect();

            // Remove old entries from both maps
            for &(old_offset, _) in &entries_to_update {
                cache.entries.remove(&old_offset);
                if let Some(access_time) = cache.access_times.remove(&old_offset) {
                    // Re-insert with updated offset
                    cache.access_times.insert(old_offset + n, access_time);
                }
            }

            // Insert updated entries
            for (old_offset, count) in entries_to_update {
                cache.entries.insert(old_offset + n, count);
            }
        }

        Ok(())
    }

    fn debug(&self) -> String {
        unsafe {
            let cache = &*self.cache.get();
            format!(
                "LruArray(cache_entries={}, cache_distance={})\n  inner: {}",
                cache.entries.len(),
                self.cache_distance,
                self.bits.debug()
            )
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::bitarray::SliceArray;

    #[test]
    fn test_lru_array_basic() {
        let inner = SliceArray::new();
        let lru = LruArray::new(inner, 4);

        // Should start empty
        assert_eq!(lru.len(), 0);
        assert_eq!(lru.cache_size(), 0);
    }

    #[test]
    fn test_lru_array_set_get() {
        let mut inner = SliceArray::new();
        inner.insert(16, 0).unwrap();

        let mut lru = LruArray::new(inner, 4);

        lru.set(0, true);
        lru.set(7, true);
        lru.set(15, true);

        assert_eq!(lru.get(0), true);
        assert_eq!(lru.get(7), true);
        assert_eq!(lru.get(15), true);
        assert_eq!(lru.get(1), false);
    }

    #[test]
    fn test_lru_array_count() {
        let mut inner = SliceArray::new();
        inner.insert(1024, 0).unwrap();

        // Use small cache distance for testing
        let mut lru = LruArray::with_cache_distance(inner, 4, 50);

        // Set some bits
        lru.set(0, true);
        lru.set(100, true);
        lru.set(200, true);
        lru.set(500, true);

        assert_eq!(lru.count(0, 100), 1);
        assert_eq!(lru.count(0, 101), 2);
        assert_eq!(lru.count(0, 500), 3);
        assert_eq!(lru.count(100, 500), 2);

        // These operations should populate the cache
        assert!(lru.cache_size() > 0);
    }

    #[test]
    fn test_lru_array_insert() {
        let mut inner = SliceArray::new();
        inner.insert(16, 0).unwrap();

        let mut lru = LruArray::new(inner, 4);

        lru.set(0, true);
        lru.set(8, true);

        // Insert 8 bits at position 4
        lru.insert(8, 4).unwrap();

        assert_eq!(lru.len(), 24);
        assert_eq!(lru.get(0), true);
        assert_eq!(lru.get(16), true); // Was at 8, now at 16
    }

    #[test]
    fn test_lru_cache_eviction() {
        let mut inner = SliceArray::new();
        inner.insert(2048, 0).unwrap();

        // Small cache size to test eviction
        let mut lru = LruArray::with_cache_distance(inner, 2, 50);

        // Set bits to trigger caching
        for i in 0..20 {
            lru.set(i * 100, true);
        }

        // Count operations that are far apart should populate cache
        lru.count(0, 500);
        lru.count(0, 1000);
        lru.count(0, 1500);

        // Cache should not exceed max size
        assert!(lru.cache_size() <= 2);
    }

    #[test]
    fn test_lru_array_total() {
        let mut inner = SliceArray::new();
        inner.insert(64, 0).unwrap();

        let mut lru = LruArray::new(inner, 4);

        lru.set(0, true);
        lru.set(10, true);
        lru.set(20, true);

        assert_eq!(lru.total(), 3);
    }
}
