package msgparser

import (
	"fmt"
	"github.com/nyan233/littlerpc/protocol/message/gen"
	"testing"
)

func BenchmarkParser(b *testing.B) {
	parser := Get(DefaultParser)(NewDefaultSimpleAllocTor(), MaxBufferSize)
	for i := 1; i <= (1 << 10); i *= 4 {
		b.Run(fmt.Sprintf("LRPCProtocol-OneParse-%dMessage", i), func(b *testing.B) {
			b.StopTimer()
			var runCount int
			parser.Reset()
			messages := make([]byte, 0, i*64)
			for j := 0; j < i; j++ {
				messages = append(messages, gen.NoMuxToBytes(gen.Big)...)
			}
			b.StartTimer()
			b.ReportAllocs()
			for j := 0; j < b.N; j++ {
				parseMsgs, err := parser.Parse(messages)
				if err != nil {
					_, err = parser.Parse(messages)
					b.Fatal(err)
				}
				b.StopTimer()
				for _, v := range parseMsgs {
					parser.Free(v.Message)
				}
				b.StartTimer()
				b.SetBytes(int64(len(messages)))
				runCount++
			}
			b.ReportMetric(float64(runCount), "RunCount")
		})
	}
}
