use std::fmt;

#[derive(Debug, Clone, PartialEq)]
pub enum K2TreeError {
    OutOfBounds {
        index: usize,
        length: usize,
    },
    InvalidOffset {
        offset: usize,
        length: usize,
    },
    InvalidSize {
        size: usize,
        reason: String,
    },
    InsertError(String),
}

impl fmt::Display for K2TreeError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            K2TreeError::OutOfBounds { index, length } => {
                write!(f, "Index {} is out of bounds (length: {})", index, length)
            }
            K2TreeError::InvalidOffset { offset, length } => {
                write!(
                    f,
                    "Invalid offset {} for array of length {}",
                    offset, length
                )
            }
            K2TreeError::InvalidSize { size, reason } => {
                write!(f, "Invalid size {}: {}", size, reason)
            }
            K2TreeError::InsertError(msg) => {
                write!(f, "Insert error: {}", msg)
            }
        }
    }
}

impl std::error::Error for K2TreeError {}
