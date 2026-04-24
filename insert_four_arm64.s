// +build arm64,!gccgo,!appengine

#include "textflag.h"

// func insertFourBitsAsm(src *byte, len uint64, in byte) (ret byte)
//
// Processes 16 bytes per iteration using NEON. For each iteration on
// block V0 of 16 bytes with incoming carry byte C (high nibble only):
//   result[i] = (V0[i] >> 4) | (V0[i-1] << 4)    for i in 1..15
//   result[0] = (V0[0] >> 4) | C
//   new C    = (V0[15] << 4) & 0xF0
//
// V3 is used as a one-byte "carry register": byte 15 holds the carry
// between iterations so VEXT can slide it in without a round-trip to
// scalar. V0..V2 are scratch.
TEXT ·insertFourBitsAsm(SB), NOSPLIT, $0-25
	MOVD	src+0(FP), R0
	MOVD	len+8(FP), R1
	MOVBU	in+16(FP), R2
	AND	$0xF0, R2, R2

	CMP	$16, R1
	BLO	tail

	// V3 = 0 with carry placed in byte 15.
	VEOR	V3.B16, V3.B16, V3.B16
	VMOV	R2, V3.B[15]

mainLoop:
	VLD1	(R0), [V0.B16]
	VUSHR	$4, V0.B16, V1.B16	// V1[i] = V0[i] >> 4  (byte-wise)
	VSHL	$4, V0.B16, V0.B16	// V0[i] = V0[i] << 4  (byte-wise)
	// V2 = [V3[15], V0[0], V0[1], ..., V0[14]]
	VEXT	$15, V0.B16, V3.B16, V2.B16
	VORR	V1.B16, V2.B16, V2.B16
	// Stash V0[15] as the next carry in V3[15].
	VMOV	V0.B[15], V3.B[15]
	VST1	[V2.B16], (R0)

	ADD	$16, R0, R0
	SUB	$16, R1, R1
	CMP	$16, R1
	BHS	mainLoop

	// Move the carry back to scalar for the tail / return.
	VMOV	V3.B[15], R2

tail:
	CBZ	R1, done

tailLoop:
	MOVBU	(R0), R3
	LSR	$4, R3, R4
	ORR	R2, R4, R4
	MOVB	R4, (R0)
	LSL	$4, R3, R2
	AND	$0xF0, R2, R2
	ADD	$1, R0, R0
	SUB	$1, R1, R1
	CBNZ	R1, tailLoop

done:
	MOVB	R2, ret+24(FP)
	RET
