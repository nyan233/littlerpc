package metadata

import (
	"github.com/nyan233/littlerpc/core/common/metadata"
)

func InputOffset(m *metadata.Process) int {
	// NOTE : 2025/07/28 默认为1, 首个参数必定为*context.Context
	return 1
}
