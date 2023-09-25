#include  "textflag.h"

TEXT    ·cmn(SB), NOSPLIT, $0
  MOVB    cmn+0(FP), AX
  MOVW    $0x29, BX
  CMPB    AX, BX
  JNE     cmnn
  MOVW    $0xef, r+16(FP)
  RET
cmnn:
  MOVW    $0x00, ret+16(FP)
  RET

TEXT    ·vs(SB), NOSPLIT, $0
  MOVB     vs+0(FP), AX
  SHRB      $4, AX
  MOVB     AX, ret+16(FP)  
  RET

TEXT    ·msgt(SB), NOSPLIT, $0
  MOVB     msgt+1(FP), AX
  MOVB     AX, ret+16(FP)  
  RET

TEXT    ·smt(SB), NOSPLIT, $0
  MOVQ     smt+0(FP), BX
  MOVB     mt+8(FP), AX
  MOVB     AX, 1(BX)
  RET

TEXT    ·spe(SB), NOSPLIT, $0
  MOVQ     smt+0(FP), BX
  MOVB     $0xff, AX
  MOVB     AX, 1(BX)
  RET
