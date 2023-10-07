#include  "textflag.h"

TEXT    ·smn(SB), NOSPLIT, $0
  MOVQ     smn+0(FP), BX
  MOVW     $0x29, AX
  MOVB     AX, 0(BX)
  RET

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

TEXT    ·iR(SB), NOSPLIT, $0
  MOVB    iR+2(FP), AX
  ANDB    $0x01, AX
  MOVW    $0x01, BX
  CMPB    AX, BX
  JNE     iRn
  MOVW    $0x01, r+16(FP)
  RET
iRn:
  MOVW    $0x00, r+16(FP)
  RET

TEXT    ·sRt(SB), NOSPLIT, $0
  MOVQ     sRt+0(FP), BX
  MOVQ     $0x01, AX
  ORB      AX, 2(BX)
  RET

TEXT    ·sRf(SB), NOSPLIT, $0
  MOVQ     sRt+0(FP), BX
  MOVQ     $0xfe, AX
  ANDB     AX, 2(BX)
  RET

TEXT    ·ivt(SB), NOSPLIT, $0
  MOVB    ivt+2(FP), AX
  ANDB    $0x02, AX
  MOVW    $0x02, BX
  CMPB    AX, BX
  JNE     ivtn
  MOVW    $0x01, r+16(FP)
  RET
ivtn:
  MOVW    $0x00, r+16(FP)
  RET

TEXT    ·svt(SB), NOSPLIT, $0
  MOVQ     svt+0(FP), BX
  MOVQ     $0x02, AX
  ORB      AX, 2(BX)
  RET

TEXT    ·mrm(SB), NOSPLIT, $0
  MOVB    mrm+2(FP), AX
  ANDB    $0x04, AX
  SHRB    $2, AX
  MOVB    AX, r+16(FP)
  RET

TEXT    ·smrmg(SB), NOSPLIT, $0
  MOVQ    smrmg+0(FP), BX
  MOVQ    $0xfb, AX
  ANDB    AX, 2(BX)
  RET

TEXT    ·smrms(SB), NOSPLIT, $0
  MOVQ     smrms+0(FP), BX
  MOVQ     $0x04, AX
  ORB      AX, 2(BX)
  RET

TEXT    ·pl(SB), NOSPLIT, $0
  MOVB     pl+2(FP), AX
  SHRB     $6, AX
  MOVB     AX, ret+16(FP)  
  RET

TEXT    ·spl(SB), NOSPLIT, $0
  MOVQ     spl+0(FP), BX
  MOVB     spl+8(FP), DX
  MOVQ     $0x3f, AX
  ANDB     AX, 2(BX)
  SHLB     $6, DX
  MOVQ     $0xc0, AX
  ANDB     AX, DX
  ORB      DX, 2(BX) 
  RET

TEXT    ·st(SB), NOSPLIT, $0
  MOVQ     st+3(FP), AX
  MOVB     AX, r+16(FP)
  RET

TEXT    ·sst(SB), NOSPLIT, $0
  MOVQ     sst+0(FP), AX
  MOVQ     sst+8(FP), BX
  MOVB     BX, 3(AX)
  RET

TEXT    ·sn(SB), NOSPLIT, $0
  MOVB     sn+4(FP), AX
  MOVB     sn+5(FP), BX
  MOVB     sn+6(FP), CX
  MOVB     sn+7(FP), DX
  SHLQ     $8, AX
  ORB      BX, AX
  SHLQ     $8, AX
  ORB      CX, AX
  SHLQ     $8, AX
  ORB      DX, AX
  MOVQ     AX, ret+16(FP)
  RET

TEXT    ·ssn(SB), NOSPLIT, $0
  MOVQ     ssn+0(FP), AX
  MOVQ     ssn+8(FP), BX
  MOVQ     BX, CX
  ANDQ     $0xff ,BX
  MOVB     BX, 7(AX)
  MOVQ     CX, BX
  SHRQ     $8, BX
  ANDQ     $0xff ,BX
  MOVB     BX, 6(AX)
  MOVQ     CX, BX
  SHRQ     $16, BX
  ANDQ     $0xff ,BX
  MOVB     BX, 5(AX)
  MOVQ     CX, BX
  SHRQ     $24, BX
  ANDQ     $0xff ,BX
  MOVB     BX, 4(AX)
  RET

TEXT    ·tid(SB), NOSPLIT, $0
  MOVB     tid+8(FP), AX
  MOVB     tid+9(FP), BX
  SHLW     $8, AX
  ORB      BX, AX
  MOVW     AX, ret+16(FP)
  RET

TEXT    ·stid(SB), NOSPLIT, $0
  MOVQ     ssn+0(FP), AX
  MOVQ     ssn+8(FP), BX
  MOVQ     BX, CX
  ANDQ     $0xff ,BX
  MOVB     BX, 9(AX)
  MOVQ     CX, BX
  SHRQ     $8, BX
  ANDQ     $0xff ,BX
  MOVB     BX, 8(AX)
  RET
