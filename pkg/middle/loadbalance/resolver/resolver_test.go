package resolver

import (
	"net"
	"net/http"
	"os"
	"reflect"
	"testing"
)

const fileData = "127.0.0.1\n192.168.1.1\n192.168.1.2\n192.168.1.3\n192.168.1.4"
const httpData = fileData

var eqData = [...]string{"127.0.0.1", "192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4"}

func TestAllResolver(t *testing.T) {
	file, err := os.OpenFile("./addrs.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove("./addrs.txt")
	}()
	_, err = file.Write([]byte(fileData))
	if err != nil {
		t.Fatal(err)
	}
	http.HandleFunc("/addrs.txt", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte(httpData))
		if err != nil {
			t.Fatal(err)
		}
	})
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	go http.Serve(listener, http.DefaultServeMux)
	tmp, _ := resolverCollection.LoadOk("live")
	lrb := tmp.(*liveResolver)
	addrs, err := lrb.Instance().Parse("live://127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(addrs, []string{"127.0.0.1"}) {
		t.Error("no equal")
	}
	addrs, err = lrb.Instance().Parse("live://127.0.0.1;192.168.1.1;192.168.1.2")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(addrs, []string{"127.0.0.1", "192.168.1.1", "192.168.1.2"}) {
		t.Error("no equal")
	}
	tmp, _ = resolverCollection.LoadOk("file")
	frb := tmp.(*fileResolver)
	addrs, err = frb.Instance().Parse("file://./addrs.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(addrs, eqData[:]) {
		t.Error("no equal")
	}
	tmp, _ = resolverCollection.LoadOk("http")
	hrb := tmp.(*httpResolver)
	addrs, err = hrb.Instance().Parse("http://127.0.0.1:8080/addrs.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(addrs, eqData[:]) {
		t.Error("no equal")
	}
}

func testReslover(t *testing.T, resolver Resolver) {

}
