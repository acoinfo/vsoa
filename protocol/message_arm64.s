#include "textflag.h"

TEXT    ·cmn(SB), NOSPLIT, $0
  MOVB     cmn+0(FP), R0
  MOVW     $0x29, R1
  CMP      R1, R0
  BEQ      cmny
  MOVW     $0x00, ret+16(FP)  
  RET
cmny:
  MOVW     $0xef, R3
  MOVW     R3, ret+16(FP)
  RET

TEXT    ·vs(SB), NOSPLIT, $0
  MOVB     vs+0(FP), R1
  LSR      $4, R1
  MOVB     R1, ret+16(FP)  
  RET

TEXT    ·msgt(SB), NOSPLIT, $0
  MOVB     msgt+1(FP), R1
  MOVB     R1, ret+16(FP)  
  RET

TEXT    ·smt(SB), NOSPLIT, $0
  MOVD     smt+0(FP), R0
  MOVB     mt+8(FP), R2
  MOVB     R2, 1(R0)
  RET

TEXT    ·spe(SB), NOSPLIT, $0
  MOVD     smt+0(FP), R0
  MOVW     $0xff, R1
  MOVB     R1, 1(R0)
  RET

TEXT    ·ir(SB), NOSPLIT, $0
  MOVB     ir+1(FP), R0
  MOVW     $0x01, R1
  CMP      R1, R0
  BEQ      iry
  MOVW     $0x00, ret+16(FP)  
  RET
iry:
  MOVW     $0x01, R3
  MOVW     R3, ret+16(FP)
  RET

TEXT    ·ipe(SB), NOSPLIT, $0
  MOVB     ipe+1(FP), R0
  MOVW     $0xff, R1
  CMP      R1, R0
  BEQ      ipey
  MOVW     $0x00, ret+16(FP)  
  RET
ipey:
  MOVW     $0x01, R3
  MOVW     R3, ret+16(FP)
  RET

TEXT    ·isi(SB), NOSPLIT, $0
  MOVB     isi+1(FP), R0
  MOVW     $0x00, R1
  CMP      R1, R0
  BEQ      isiy
  MOVW     $0x00, ret+16(FP)  
  RET
isiy:
  MOVW     $0x01, R3
  MOVW     R3, ret+16(FP)
  RET

TEXT    ·iss(SB), NOSPLIT, $0
  MOVB     iss+1(FP), R0
  MOVW     $0x02, R1
  CMP      R1, R0
  BEQ      issy
  MOVW     $0x00, ret+16(FP)  
  RET
issy:
  MOVW     $0x01, R3
  MOVW     R3, ret+16(FP)
  RET

TEXT    ·ius(SB), NOSPLIT, $0
  MOVB     ius+1(FP), R0
  MOVW     $0x03, R1
  CMP      R1, R0
  BEQ      iusy
  MOVW     $0x00, ret+16(FP)  
  RET
iusy:
  MOVW     $0x01, R3
  MOVW     R3, ret+16(FP)
  RET

TEXT    ·id(SB), NOSPLIT, $0
  MOVB     id+1(FP), R0
  MOVW     $0x05, R1
  CMP      R1, R0
  BEQ      idy
  MOVW     $0x00, ret+16(FP)  
  RET
idy:
  MOVW     $0x01, R3
  MOVW     R3, ret+16(FP)
  RET

TEXT    ·ip(SB), NOSPLIT, $0
  MOVB     ip+1(FP), R0
  MOVW     $0x04, R1
  CMP      R1, R0
  BEQ      ipy
  MOVW     $0x00, ret+16(FP)  
  RET
ipy:
  MOVW     $0x01, R3
  MOVW     R3, ret+16(FP)
  RET
