package packet

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

type GzipPacket struct {}


func (g *GzipPacket) Scheme() string {
	return "gzip"
}

func (g *GzipPacket) UnPacket(p []byte) ([]byte,error) {
	gr,err := gzip.NewReader(bytes.NewReader(p))
	if err != nil {
		return nil,err
	}
	defer gr.Close()
	p,err = ioutil.ReadAll(gr)
	if err.Error() == "unexpected EOF" {
		return p,nil
	} else {
		return p,err
	}
}

func (g *GzipPacket) EnPacket(p []byte) ([]byte,error) {
	var bb bytes.Buffer
	gw, _ := gzip.NewWriterLevel(&bb,gzip.BestCompression)
	defer gw.Close()
	_, err := gw.Write(p)
	if err != nil {
		return nil, err
	}
	gw.Flush()
	return bb.Bytes(),nil
}

