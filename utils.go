package littlerpc

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/lesismal/nbio"
	"github.com/nyan233/littlerpc/coder"
	lreflect "github.com/nyan233/littlerpc/reflect"
	"reflect"
	"runtime"
	"strconv"
	"unsafe"
)

type eface struct {
	typ unsafe.Pointer
	data unsafe.Pointer
}

func checkCoderType(callerMd coder.CallerMd) (interface{},error) {
	switch callerMd.ArgType {
	// 处理额外的指针类型
	case coder.Pointer:
		var any coder.AnyArgs
		err := json.Unmarshal(callerMd.Req,&any)
		if err != nil {
			return nil, err
		}
		// encoding/json默认识别的类型可能有误，需要修复类型
		any.Any = fixJsonType(any.Any,callerMd.AppendType)
		typEface, err := mappingReflectPtrType(callerMd.AppendType)
		// 简单基础类型如int这类的和map等复杂类型处理的逻辑不一样
		if err == nil {
			// 替换类型信息
			return *(*interface{})(unsafe.Pointer(&eface{
				typ: (*eface)(unsafe.Pointer(&typEface)).typ,
				data: (*eface)(unsafe.Pointer(&any.Any)).data,
			})),nil
		}
		// 复杂类型直接使用encoding/json生成的类型信息
		return any.Any,nil
	case coder.String:
		var tmp coder.AnyArgs
		err := json.Unmarshal(callerMd.Req,&tmp)
		return tmp.Any,err
	case coder.Integer, coder.Long, coder.Float, coder.Double:
		var tmp coder.AnyArgs
		err := json.Unmarshal(callerMd.Req,&tmp)
		if err == nil {
			tmp.Any = fixJsonType(tmp.Any,callerMd.ArgType)
		}
		return tmp.Any,err
	case coder.Array:
		// 处理数组的附加类型
		var tmp coder.AnyArgs
		err := json.Unmarshal(callerMd.Req, &tmp)
		if err == nil {
			// []byte类型会被encoding/json编码为base64字符串，所以需要做特殊处理
			if callerMd.AppendType == coder.Byte {
				return base64.StdEncoding.DecodeString(tmp.Any.(string))
			}
			arrayType, err := mappingArrayNoPtrType(callerMd.AppendType,tmp.Any)
			if err != nil {
				return nil, err
			}
			tmp.Any = arrayType
		}
		return tmp.Any,nil
	case coder.Map:
		return nil,nil
	case coder.Struct:
		return nil,nil
	default:
		return nil,errors.New("type is not found")
	}
}

//func fixJsonArrayType(i interface{},typ coder.Type) interface{} {
//
//}

func fixJsonType(i interface{},typ coder.Type) interface{} {
	eType, err := mappingReflectNoPtrType(typ,i)
	if err != nil {
		return nil
	}
	return eType
}

func mappingArrayNoPtrType(typ coder.Type,value interface{}) (interface{},error) {
	if i := reflect.TypeOf(value).Kind(); !(i == reflect.Array || i == reflect.Slice || i == reflect.String) {
		return nil,errors.New("value type is not array/slice")
	}
	switch typ {
	// 转换为[]Byte只能是string或者一些基础类型或者除了string类型的数组
	case coder.Byte:
		_inter := interface{}([]byte{0})
		inter := (*eface)(unsafe.Pointer(&_inter))
		// type is string ?
		str,ok := value.(string)
		if ok {
			header := (*reflect.StringHeader)(unsafe.Pointer(&str))
			(*reflect.SliceHeader)(inter.data).Data = header.Data
			(*reflect.SliceHeader)(inter.data).Len = header.Len
			(*reflect.SliceHeader)(inter.data).Cap = header.Len
			return _inter,nil
		}
		// type is array ?
		if vType := reflect.TypeOf(value); vType.Kind() == reflect.Array {
			arrayLen := vType.Len()
			(*reflect.SliceHeader)(inter.data).Data = uintptr((*eface)(unsafe.Pointer(&value)).data)
			(*reflect.SliceHeader)(inter.data).Len = arrayLen
			(*reflect.SliceHeader)(inter.data).Cap = arrayLen
			return _inter,nil
		}
		ptr, length := lreflect.IdentifyTypeNoInfo(value)
		(*reflect.SliceHeader)(inter.data).Data = uintptr(ptr)
		(*reflect.SliceHeader)(inter.data).Len = length
		(*reflect.SliceHeader)(inter.data).Cap = length
		return _inter,nil
	case coder.Long:
		return interface{}(*new([]int32)),nil
	case coder.Integer:
		return interface{}(*new([]int64)),nil
	case coder.String:
		return interface{}(*new(string)),nil
	case coder.Float:
		return interface{}(*new([]float32)),nil
	case coder.Double:
		return interface{}(*new([]float64)),nil
	case coder.ULong:
		return interface{}(*new([]uint32)),nil
	case coder.UInteger:
		return interface{}(*new([]uint64)),nil
	default:
		return nil,errors.New("not support other type")
	}
}

func mappingReflectNoPtrType(typ coder.Type,value interface{}) (interface{},error) {
	switch typ {
	case coder.Long:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(int32(0))).Interface(),nil
	case coder.Integer:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(int64(0))).Interface(),nil
	case coder.String:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(*new(string))).Interface(),nil
	case coder.Float:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(float32(0))).Interface(),nil
	case coder.Double:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(float64(0))).Interface(),nil
	case coder.ULong:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(uint32(0))).Interface(),nil
	case coder.UInteger:
		return reflect.ValueOf(value).Convert(reflect.TypeOf(uint64(0))).Interface(),nil
	default:
		return nil,errors.New("not support other type")
	}
}

func mappingReflectPtrType(typ coder.Type) (interface{},error) {
	switch typ {
	case coder.Long:
		return reflect.ValueOf(new(int32)).Interface(),nil
	case coder.Integer:
		return reflect.ValueOf(new(int64)).Interface(),nil
	case coder.String:
		return reflect.ValueOf(new(string)).Interface(),nil
	case coder.Float:
		return reflect.ValueOf(new(float32)).Interface(),nil
	case coder.Double:
		return reflect.ValueOf(new(float64)).Interface(),nil
	case coder.ULong:
		return reflect.ValueOf(new(uint32)).Interface(),nil
	case coder.UInteger:
		return reflect.ValueOf(new(uint64)).Interface(),nil
	default:
		return nil,errors.New("not support other type")
	}
}

func checkIType(i interface{}) coder.Type {
	switch i.(type) {
	case int,int8,int16,int32,int64:
		return coder.Integer
	case uint,uint16,uint32,uint64,uintptr:
		return coder.UInteger
	case uint8:
		return coder.Byte
	case string:
		return coder.String
	case float32:
		return coder.Float
	case float64:
		return coder.Double
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Slice,reflect.Array:
		return coder.Array
	case reflect.Map:
		return coder.Map
	case reflect.Struct:
		return coder.Struct
	case reflect.Ptr:
		return coder.Pointer
	default:
		panic("the type error")
	}
}


func HandleError(sp coder.RStackFrame,errNo coder.Error,conn *nbio.Conn,appendInfo string,more ...interface{}) {
	md := coder.CalleeMd{
		ArgType:    coder.Struct,
		Rep:        nil,
	}
	switch errNo.Info {
	case ErrJsonUnMarshal.Info:
		_, file, line, _ := runtime.Caller(2)
		errNo.Trace = file + ":" + strconv.Itoa(line)
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Response = append(sp.Response,md)
		errNoBytes, err := json.Marshal(&sp)
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
		sp.Response = append(sp.Response,md)
		errNoBytes, err := json.Marshal(&sp)
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
		sp.Response = append(sp.Response,md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.Write(errNoBytes)
	case Nil.Info:
		err := md.EncodeResponse(errNo)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		sp.Response = append(sp.Response,md)
		errNoBytes, err := json.Marshal(&sp)
		if err != nil {
			panic(errors.New("encoding/json marshal failed"))
		}
		conn.Write(errNoBytes)
	}
}

func encodeAnyBytes() {}