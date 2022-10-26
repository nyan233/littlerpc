package metadata

import (
	"context"
	"github.com/nyan233/littlerpc/pkg/common/metadata"
	"github.com/nyan233/littlerpc/pkg/stream"
	"reflect"
)

func InputOffset(m *metadata.Process) int {
	switch {
	case m.SupportStream && m.SupportContext:
		return 2
	case m.SupportContext, m.SupportStream:
		return 1
	default:
		return 0
	}
}

// IFContextOrStream 检查输入参数中是否有context&stream
// type必须为Method/Func类型
func IFContextOrStream(opt *metadata.Process, typ reflect.Type) (offset int) {
	// j <= (method.Type.NumIn() > 2 ? 2 : method.Type.NumIn())
	// j <= (3 > 2 ? 2 : 3) --> j <= 2
	for j := 1; j <= (typ.NumIn()-1)&2; j++ {
		switch reflect.New(typ.In(j)).Interface().(type) {
		case *context.Context:
			opt.SupportContext = true
			offset++
		case *stream.LStream:
			opt.SupportStream = true
			offset++
		}
	}
	return
}
