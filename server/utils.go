package server

import (
	"encoding/json"
	"errors"
	"github.com/lesismal/nbio"
	"github.com/nyan233/littlerpc/coder"
	"runtime"
	"strconv"
)

func checkType(callerMd coder.CallerMd) (interface{},error) {
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
