package analysis

import (
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol/message"
	mux2 "github.com/nyan233/littlerpc/protocol/message/mux"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAnalysisNoMuxMessage(t *testing.T) {
	var bytes []byte
	assert.Equal(t, message.Marshal(message.New(), (*container.Slice[byte])(&bytes)), nil)
	t.Log(NoMux(bytes))
}

func TestAnalysisMuxMessage(t *testing.T) {
	var bytes []byte
	assert.Equal(t, message.Marshal(message.New(), (*container.Slice[byte])(&bytes)), nil)
	muxBlock := &mux2.Block{
		Flags:    mux2.Enabled,
		StreamId: random.FastRand(),
		MsgId:    uint64(random.FastRand()),
	}
	muxBlock.SetPayloads(bytes)
	var bytes2 []byte
	mux2.Marshal(muxBlock, (*container.Slice[byte])(&bytes2))
	t.Log(Mux(bytes2))
}
