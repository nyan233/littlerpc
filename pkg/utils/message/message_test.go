package message

import (
	"github.com/nyan233/littlerpc/pkg/container"
	"github.com/nyan233/littlerpc/pkg/utils/random"
	"github.com/nyan233/littlerpc/protocol"
	"testing"
)

func TestMessageUtils(t *testing.T) {
	var bytes []byte
	protocol.MarshalMessage(protocol.NewMessage(), (*container.Slice[byte])(&bytes))
	t.Log(AnalysisMessage(bytes))
}

func TestMuxMessageUtils(t *testing.T) {
	var bytes []byte
	protocol.MarshalMessage(protocol.NewMessage(), (*container.Slice[byte])(&bytes))
	muxBlock := &protocol.MuxBlock{
		Flags:    protocol.MuxEnabled,
		StreamId: random.FastRand(),
		MsgId:    uint64(random.FastRand()),
	}
	muxBlock.SetPayloads(bytes)
	var bytes2 []byte
	err := protocol.MarshalMuxBlock(muxBlock, (*container.Slice[byte])(&bytes2))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(AnalysisMuxMessage(bytes2))
}
