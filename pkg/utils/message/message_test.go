package message

import (
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol/message"
	"github.com/nyan233/littlerpc/protocol/mux"
	"testing"
)

func TestMessageUtils(t *testing.T) {
	var bytes []byte
	message.Marshal(message.New(), (*container.Slice[byte])(&bytes))
	t.Log(AnalysisMessage(bytes))
}

func TestMuxMessageUtils(t *testing.T) {
	var bytes []byte
	message.Marshal(message.New(), (*container.Slice[byte])(&bytes))
	muxBlock := &mux.Block{
		Flags:    mux.Enabled,
		StreamId: random.FastRand(),
		MsgId:    uint64(random.FastRand()),
	}
	muxBlock.SetPayloads(bytes)
	var bytes2 []byte
	err := mux.Marshal(muxBlock, (*container.Slice[byte])(&bytes2))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(AnalysisMuxMessage(bytes2))
}
