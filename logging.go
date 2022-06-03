package littlerpc

import (
	"github.com/zbh255/bilog"
	"os"
)

var logger bilog.Logger = bilog.NewLogger(os.Stdout,bilog.PANIC,bilog.WithTimes(),
	bilog.WithCaller(),bilog.WithLowBuffer(0),bilog.WithTopBuffer(0))
