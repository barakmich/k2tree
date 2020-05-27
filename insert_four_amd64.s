// +build amd64,!gccgo,!appengine

#include "textflag.h"

TEXT ·insertFourBitsAsm(SB), NOSPLIT, $0
	MOVQ    src+0(FP), SI
	MOVQ    len+8(FP), CX
	MOVBQZX in+16(FP), BX

	SHLQ $4, BX
	ANDQ $0xF0, BX
	MOVQ $0xF0F0F0F0F0F0F0F0, R11

	CMPQ CX, $32
	JB   tail

	//	MOVQ SI, R8
	//	ANDQ $-32, R8
	//	ADDQ $32, R8
	//	SUBQ SI, R8
	//	JZ mainLoop
	//
	// headLoop:
	//	MOVB (SI), AX
	//	MOVB AX, DX
	//	SHRB $4, AX
	//	ORB  BX, AX
	//	MOVB AX, (SI)
	//	MOVB DX, BX
	//	SHLB $4, BX
	//	LEAQ 1(SI), SI
	//	DECQ CX
	//	DECQ R8
	//	JNZ headLoop
	//
	//	CMPQ CX, $32
	//	JB tail

mainLoop:
	MOVQ (SI), AX
	MOVQ (SI), DX
	MOVQ 8(SI), R8
	MOVQ 8(SI), R9
	MOVQ 16(SI), R10
	MOVQ 16(SI), R12
	MOVQ 24(SI), R13
	MOVQ 24(SI), R14
	ANDQ R11, AX
	ANDQ R11, R8
	ANDQ R11, R10
	ANDQ R11, R13
	RORQ $4, AX
	RORQ $4, R8
	RORQ $4, R10
	RORQ $4, R13
	ROLQ $12, DX
	ROLQ $12, R9
	ROLQ $12, R12
	ROLQ $12, R14
	ANDQ R11, DX
	ANDQ R11, R9
	ANDQ R11, R12
	ANDQ R11, R14
	ORQ  DX, AX
	ORQ  R9, R8
	ORQ  R12, R10
	ORQ  R14, R13
	ANDQ $-0xF1, AX
	ANDQ $-0xF1, R8
	ANDQ $-0xF1, R10
	ANDQ $-0xF1, R13

	ORQ BX, AX

	ANDQ $0xF0, DX
	ORQ  DX, R8

	ANDQ $0xF0, R9
	ORQ  R9, R10

	ANDQ $0xF0, R12
	ORQ  R12, R13

	MOVQ AX, (SI)
	MOVQ R8, 8(SI)
	MOVQ R10, 16(SI)
	MOVQ R13, 24(SI)

	MOVQ R14, BX
	ANDQ $0xF0, BX
	LEAQ 32(SI), SI
	SUBQ $32, CX
	JZ   return

	CMPQ CX, $32
	JAE  mainLoop

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
	JNZ  tail

ret:
	MOVB BX, ret+24(FP)
	RET
