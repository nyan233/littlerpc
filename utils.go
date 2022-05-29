package littlerpc

import (
	"encoding/json"
	"errors"
	"github.com/lesismal/nbio"
	"github.com/nyan233/littlerpc/coder"
	"reflect"
	"runtime"
	"strconv"
)

func checkCoderType(callerMd coder.CallerMd) (interface{},error) {
	switch callerMd.ArgType {
	case coder.String,coder.Integer,coder.Long,coder.Float, coder.Double, coder.Array:
		var tmp coder.AnyArgs
		err := json.Unmarshal(callerMd.Req,&tmp)
		return tmp.Any,err
	case coder.Map:
		return nil,nil
	case coder.Struct:
		return nil,nil
	default:
		return nil,errors.New("type is not found")
	}
}

func checkIType(i interface{}) coder.Type {
	switch i.(type) {
	case int,int8,int16,int32,int64:
		return coder.Integer
	case uint,uint8,uint16,uint32,uint64:
		return coder.UInteger
	case string:
		return coder.String
	case float32,float64:
		return coder.Double
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Slice,reflect.Array:
		return coder.Array
	case reflect.Map:
		return coder.Map
	case reflect.Struct:
		return coder.Struct
	default:
		panic("the type error")
	}
}


func HandleError(md coder.CalleeMd,errNo coder.Error,conn *nbio.Conn,appendInfo string,more ...interface{}) {
	md.ArgType = coder.Struct
	switch errNo.Info {
	case ErrJsonUnMarshal.Info:
		_, file, line, _ := runtime.Caller(2)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		errNoBytes, err := json.Marshal(&md)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.Write(errNoBytes)
		break
	case ErrMethodNoRegister.Info:
		_, file, line, _ := runtime.Caller(2)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		errNoBytes, err := json.Marshal(&md)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.Write(errNoBytes)
		break
	case ErrServer.Info:
		errNo.Info += appendInfo
		_, file, line, _ := runtime.Caller(1)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		errNoBytes, err := json.Marshal(&md)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.Write(errNoBytes)
	case Nil.Info:
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		errNoBytes, err := json.Marshal(&md)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.Write(errNoBytes)
	}
}
