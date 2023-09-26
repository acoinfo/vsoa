#include  "textflag.h"

TEXT    ·cmn(SB), NOSPLIT, $0
  MOVB    cmn+0(FP), AX
  MOVW    $0x29, BX
  CMPB    AX, BX
  JNE     cmnn
  MOVW    $0xef, r+16(FP)
  RET
cmnn:
  MOVW    $0x00, r+16(FP)
  RET

TEXT    ·vs(SB), NOSPLIT, $0
  MOVB     vs+0(FP), AX
  SHRB     $4, AX
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

TEXT    ·ir(SB), NOSPLIT, $0
  MOVB    cmn+1(FP), AX
  MOVW    $0x01, BX
  CMPB    AX, BX
  JNE     irn
  MOVW    $0x01, r+16(FP)
  RET
irn:
  MOVW    $0x00, r+16(FP)
  RET

TEXT    ·ipe(SB), NOSPLIT, $0
  MOVB    ipe+1(FP), AX
  MOVW    $0xff, BX
  CMPB    AX, BX
  JNE     ipen
  MOVW    $0x01, r+16(FP)
  RET
ipen:
  MOVW    $0x00, r+16(FP)
  RET

TEXT    ·isi(SB), NOSPLIT, $0
  MOVB    isi+1(FP), AX
  MOVW    $0x00, BX
  CMPB    AX, BX
  JNE     isin
  MOVW    $0x01, r+16(FP)
  RET
isin:
  MOVW    $0x00, r+16(FP)
  RET

TEXT    ·iss(SB), NOSPLIT, $0
  MOVB    iss+1(FP), AX
  MOVW    $0x02, BX
  CMPB    AX, BX
  JNE     issn
  MOVW    $0x01, r+16(FP)
  RET
issn:
  MOVW    $0x00, r+16(FP)
  RET

TEXT    ·ius(SB), NOSPLIT, $0
  MOVB    ius+1(FP), AX
  MOVW    $0x03, BX
  CMPB    AX, BX
  JNE     iusn
  MOVW    $0x01, r+16(FP)
  RET
iusn:
  MOVW    $0x00, r+16(FP)
  RET

TEXT    ·id(SB), NOSPLIT, $0
  MOVB    id+1(FP), AX
  MOVW    $0x05, BX
  CMPB    AX, BX
  JNE     idn
  MOVW    $0x01, r+16(FP)
  RET
idn:
  MOVW    $0x00, r+16(FP)
  RET

TEXT    ·ip(SB), NOSPLIT, $0
  MOVB    ip+1(FP), AX
  MOVW    $0x04, BX
  CMPB    AX, BX
  JNE     ipn
  MOVW    $0x01, r+16(FP)
  RET
ipn:
  MOVW    $0x00, r+16(FP)
  RET

