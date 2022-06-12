package packet

import "testing"

func TestGzip(t *testing.T) {
	gz := &GzipPacket{}
	bigBytes := make([]byte,1 << 20)
	initStr := "hello world"
	for i := 0; i < len(bigBytes); i += len(initStr){
		copy(bigBytes[i:],initStr)
	}
	bytes, err := gz.EnPacket(bigBytes)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err = gz.UnPacket(bytes)
	if err != nil {
		t.Fatal(err)
	}
}

func TestText(t *testing.T) {
	text := &TextPacket{}
	bigBytes := make([]byte,1 << 20)
	initStr := "hello world"
	for i := 0; i < len(bigBytes); i += len(initStr){
		copy(bigBytes[i:],initStr)
	}
	bytes, err := text.EnPacket(bigBytes)
	if err != nil {
		t.Fatal(err)
	}
	bytes, err = text.UnPacket(bytes)
	if err != nil {
		t.Fatal(err)
	}
}