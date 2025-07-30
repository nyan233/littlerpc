package main

import (
	"fmt"
	"strings"
)

type Argument struct {
	Name string
	Type string
}

type Statement struct {
	ReceiveName string
	CallFunc    string
	InputList   []string
	OutputList  []string
}

type Method struct {
	Receive     Argument
	ServiceName string
	Name        string
	InputList   []Argument
	OutputList  []Argument
	Statement   Statement
}

func (m *Method) FormatToSync() string {
	var (
		recvName          = m.Receive.Name
		recvType          = m.Receive.Type
		methodName        = m.Name
		argList           strings.Builder
		returnList        strings.Builder
		rpcRequestArgList strings.Builder
		callNewObjList    strings.Builder
		fmtStr            = `func (%s %s) %s (%sopts ...client.CallOption) (%s) {%sr%d = %s.Request2("%s", opts, %d, %s);return}`
	)
	for index, input := range m.InputList {
		fmt.Fprintf(&argList, "a%d %s", index, input.Type)
		fmt.Fprintf(&rpcRequestArgList, "a%d, ", index)
		if index != len(m.InputList)-1 || len(m.OutputList) > 0 {
			argList.WriteString(", ")
		}
	}
	for index, output := range m.OutputList {
		fmt.Fprintf(&returnList, "r%d %s", index, output.Type)
		if output.Type != "error" {
			if strings.HasPrefix(output.Type, "*") {
				fmt.Fprintf(&callNewObjList, "r%d = new(%s);", index, output.Type[1:])
				fmt.Fprintf(&rpcRequestArgList, "r%d, ", index)
			} else {
				fmt.Fprintf(&rpcRequestArgList, "&r%d, ", index)
			}
		}
		if index != len(m.OutputList)-1 {
			returnList.WriteString(", ")
		}
	}
	return fmt.Sprintf(fmtStr, recvName, recvType, methodName, argList.String(), returnList.String(), callNewObjList.String(), len(m.OutputList)-1, recvName, m.ServiceName, len(m.InputList), rpcRequestArgList.String())
	//var sb strings.Builder
	//_, _ = fmt.Fprintf(&sb, "func(%s %s) %s", m.Receive.Name, m.Receive.Type, m.Name)
	//sb2 := strings.Builder{}
	//inputNames := strings.Builder{}
	//for _, arg := range m.InputList {
	//	_, _ = fmt.Fprintf(&sb2, "%s %s,", arg.Name, arg.Type)
	//	_, _ = fmt.Fprintf(&inputNames, "%s,", arg.Name)
	//}
	//_, _ = fmt.Fprintf(&sb, "(%sopts ...client.CallOption)", sb2.String())
	//sb2.Reset()
	// var nop string
	//if len(m.OutputList) > 1 {
	//	nop = ","
	//}
	//for _, arg := range m.OutputList {
	//	if arg.Name == "" {
	//		_, _ = fmt.Fprintf(&sb2, "%s%s", arg.Type, nop)
	//	} else {
	//		_, _ = fmt.Fprintf(&sb2, "%s %s%s", arg.Name, arg.Type, nop)
	//	}
	//}
	//if len(m.OutputList) > 1 {
	//	_, _ = fmt.Fprintf(&sb, "(%s)", sb2.String())
	//} else {
	//	_, _ = fmt.Fprintf(&sb, "%s", sb2.String())
	//}
	//var outNames strings.Builder
	//var assertSet strings.Builder
	//for i := 0; i < len(m.OutputList)-1; i++ {
	//	_, _ = fmt.Fprintf(&assertSet, "r%d,_ := reps[%d].(%s);", i, i, m.OutputList[i].Type)
	//	_, _ = fmt.Fprintf(&outNames, "r%d,", i)
	//}
	//if len(m.OutputList) == 1 {
	//	_, _ = fmt.Fprintf(&sb, "{_,err := %s.Call(\"%s\",opts,%s);%sreturn %serr}", m.Receive.Name, m.ServiceName,
	//		inputNames.String(), assertSet.String(), outNames.String())
	//} else if len(m.OutputList) > 1 {
	//	_, _ = fmt.Fprintf(&sb, "{reps,err := %s.Call(\"%s\",opts,%s);%sreturn %serr}", m.Receive.Name, m.ServiceName,
	//		inputNames.String(), assertSet.String(), outNames.String())
	//}
	//return sb.String()

}

func (m *Method) FormatToASync() string {
	return ""
}

func (m *Method) FormatToRequests() string {
	return ""
}
