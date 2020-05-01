// +build amd64,!gccgo,!appengine

#include "textflag.h"


TEXT ·insertFourBitsAsm(SB),NOSPLIT,$0
	MOVQ src+0(FP), SI
	MOVQ len+8(FP), CX
	MOVBQZX in+16(FP), BX

	SHLB $4, BX
	MOVQ $0xF0F0F0F0F0F0F0F0, R11

	CMPQ CX, $8
	JB tail

eightloop:
	MOVQ (SI), AX
	MOVQ AX, DX
	ANDQ R11, AX
	RORQ $4, AX
	ROLQ $12, DX
	ANDQ R11, DX
	ORQ  DX, AX
	ANDQ $-0xF1, AX
	ORB  BX, AX
	MOVQ AX, (SI)
	MOVQ DX, BX
	ANDQ $0xF0, BX
	LEAQ 8(SI), SI
	SUBQ $8, CX
	JZ ret

	CMPQ CX, $8
	JAE eightloop

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
