package analysis

import (
	"github.com/nyan233/littlerpc/core/container"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/protocol/message/mux"
	"github.com/nyan233/littlerpc/core/utils/random"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAnalysisNoMuxMessage(t *testing.T) {
	var bytes []byte
	assert.Equal(t, message2.Marshal(message2.New(), (*container.Slice[byte])(&bytes)), nil)
	t.Log(NoMux(bytes))
}

func TestAnalysisMuxMessage(t *testing.T) {
	var bytes []byte
	assert.Equal(t, message2.Marshal(message2.New(), (*container.Slice[byte])(&bytes)), nil)
	muxBlock := &mux.Block{
		Flags:    mux.Enabled,
		StreamId: random.FastRand(),
		MsgId:    uint64(random.FastRand()),
	}
	muxBlock.SetPayloads(bytes)
	var bytes2 []byte
	mux.Marshal(muxBlock, (*container.Slice[byte])(&bytes2))
	t.Log(Mux(bytes2))
}
