package random

import (
	"github.com/nyan233/littlerpc/protocol/message"
	"math/rand"
	"strings"
	"time"
	"unsafe"
)

func GenStringOnAscii(maxBytes uint32) string {
	nByte := int(FastRandN(maxBytes))
	var sb strings.Builder
	sb.Grow(nByte)
	for i := 0; i < nByte; i++ {
		sb.WriteByte(byte(FastRandN(26)) + 65)
	}
	return sb.String()
}

func GenStringsOnAscii(maxNStr, maxBytes uint32) []string {
	nStr := int(FastRandN(maxNStr))
	strs := make([]string, nStr)
	for i := 0; i < nStr; i++ {
		strs[i] = GenStringOnAscii(maxBytes)
	}
	return strs
}

func GenBytesOnAscii(maxBytes uint32) []byte {
	str := GenStringOnAscii(maxBytes)
	return *(*[]byte)(unsafe.Pointer(&str))
}

func GenSequenceNumberOnMathRand(nSeq int) []uint32 {
	seq := make([]uint32, nSeq)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < nSeq; i++ {
		seq[i] = rand.Uint32()
	}
	return seq
}

func GenSequenceNumberOnFastRand(nSeq int) []uint32 {
	seq := make([]uint32, nSeq)
	for i := 0; i < nSeq; i++ {
		seq[i] = FastRand()
	}
	return seq
}

func GenProtocolMessage() *message.Message {
	msg := message.New()
	msg.SetMsgId(uint64(FastRand()))
	msg.SetCodecType(uint8(FastRand()))
	msg.SetEncoderType(uint8(FastRand()))
	msg.SetMsgType(uint8(FastRand()))
	msg.SetInstanceName(GenStringOnAscii(100))
	msg.SetMethodName(GenStringOnAscii(100))
	for i := 0; i < int(FastRandN(10)+1); i++ {
		msg.AppendPayloads(GenBytesOnAscii(FastRandN(50)))
	}
	for i := 0; i < int(FastRandN(10)+1); i++ {
		msg.MetaData.Store(GenStringOnAscii(10), GenStringOnAscii(10))
	}
	return msg
}
