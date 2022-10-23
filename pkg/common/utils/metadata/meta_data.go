package metadata

import "github.com/nyan233/littlerpc/pkg/common"

func InputOffset(m *common.Method) int {
	switch {
	case m.SupportStream && m.SupportContext:
		return 2
	case m.SupportContext, m.SupportStream:
		return 1
	default:
		return 0
	}
}
