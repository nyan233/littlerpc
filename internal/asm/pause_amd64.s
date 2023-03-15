//go:build amd64

#include "textflag.h"

TEXT ·PAUSE(SB),NOSPLIT,$0
    PAUSE
    RET