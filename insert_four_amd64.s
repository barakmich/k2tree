// +build amd64,!gccgo,!appengine

#include "textflag.h"


TEXT ·insertFourBitsAsm(SB),NOSPLIT,$0
	MOVQ src+0(FP), SI
	MOVQ len+8(FP), CX
	MOVBQZX in+16(FP), BX

	SHLQ $4, BX
	ANDQ $0xF0, BX
	MOVQ $0xF0F0F0F0F0F0F0F0, R11

	CMPQ CX, $16
	JB tail

//        ;MOVQ SI, R9
//	;ANDQ $7, R9
//	;JZ sixteenLoop
//	;SUBQ $8, R9
//
//;head:
//	;MOVB (SI), AX
//	;MOVB AX, DX
//	;SHRB $4, AX
//	;ORB  BX, AX
//	;MOVB AX, (SI)
//	;MOVB DX, BX
//	;SHLB $4, BX
//	;LEAQ 1(SI), SI
//	;DECQ CX
//	;DECQ R9
//	;JNZ head

sixteenLoop:
	MOVQ (SI), AX
	MOVQ 8(SI), R8
	MOVQ AX, DX
	MOVQ R8, R9
	ANDQ R11, AX
	ANDQ R11, R8
	RORQ $4, AX
	RORQ $4, R8
	ROLQ $12, DX
	ROLQ $12, R9
	ANDQ R11, DX
	ANDQ R11, R9
	ORQ  DX, AX
	ORQ  R9, R8
	ANDQ $-0xF1, AX
	ANDQ $-0xF1, R8
	ORQ  BX, AX
	MOVQ AX, (SI)
	MOVQ DX, R10
	ANDQ $0xF0, R10
	ORQ  R10, R8
	MOVQ R8, 8(SI)
	MOVQ R9, BX
	ANDQ $0xF0, BX
	LEAQ 16(SI), SI
	SUBQ $16, CX
	JZ ret

	CMPQ CX, $16
	JAE sixteenLoop

tail:
	MOVB (SI), AX
	MOVB AX, DX
	SHRB $4, AX
	ORB  BX, AX
	MOVB AX, (SI)
	MOVB DX, BX
	SHLB $4, BX
	LEAQ 1(SI), SI
	DECQ CX
	JNZ tail
ret:
	MOVB BX, ret+24(FP)
	RET
