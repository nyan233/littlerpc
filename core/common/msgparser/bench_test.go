package msgparser

import (
	"fmt"
	"github.com/nyan233/littlerpc/core/container"
	message2 "github.com/nyan233/littlerpc/core/protocol/message"
	"github.com/nyan233/littlerpc/core/protocol/message/gen"
	"io"
	"testing"
)

func BenchmarkParser(b *testing.B) {
	const (
		UsageParseOnReader = true
	)
	parser := Get(DefaultParser)(NewDefaultSimpleAllocTor(), MaxBufferSize*32)
	for i := 1; i <= (1 << 10); i *= 4 {
		b.Run(fmt.Sprintf("LRPCProtocol-OneParse-%dMessage", i), func(b *testing.B) {
			b.StopTimer()
			var runCount int
			parser.Reset()
			messages := make([]byte, 0, i*64)
			messageSplits := make([]*message2.Message, 0)
			lengths := make([]int, 0)
			for j := 0; j < i; j++ {
				msg := gen.NoMux(gen.Little)
				messageSplits = append(messageSplits, msg)
				var bytes container.Slice[byte]
				err := message2.Marshal(msg, &bytes)
				if err != nil {
					b.Fatal(err)
				}
				lengths = append(lengths, bytes.Len())
				messages = append(messages, bytes...)
			}
			var point int
			for index, length := range lengths {
				msg := message2.New()
				err := message2.Unmarshal(messages[point:point+length], msg)
				if err != nil {
					var bytes container.Slice[byte]
					err = message2.Marshal(messageSplits[index], &bytes)
					b.Fatal(index, length, err)
				}
				point += length
			}
			buf2 := make([]byte, len(messages))
			b.StartTimer()
			b.ReportAllocs()
			for j := 0; j < b.N; j++ {
				var parseMsgs []ParserMessage
				var err error
				if UsageParseOnReader {
					parseMsgs, err = parser.ParseOnReader(func(bytes []byte) (n int, err error) {
						return copy(bytes, messages), io.EOF
					})
				} else {
					// 模拟网络框架的拷贝
					copy(buf2, messages)
					parseMsgs, err = parser.Parse(buf2)
				}
				if err != nil {
					_, err = parser.Parse(messages)
					b.Fatal(j, err)
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
