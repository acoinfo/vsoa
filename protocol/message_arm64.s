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
  MOVB     $0xff, R1
  MOVB     R1, 1(R0)
  RET
