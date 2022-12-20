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
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "func(%s %s) %s", m.Receive.Name, m.Receive.Type, m.Name)
	sb2 := strings.Builder{}
	inputNames := strings.Builder{}
	for _, arg := range m.InputList {
		_, _ = fmt.Fprintf(&sb2, "%s %s,", arg.Name, arg.Type)
		_, _ = fmt.Fprintf(&inputNames, "%s,", arg.Name)
	}
	_, _ = fmt.Fprintf(&sb, "(%s)", sb2.String())
	sb2.Reset()
	var nop string
	if len(m.OutputList) > 1 {
		nop = ","
	}
	for _, arg := range m.OutputList {
		if arg.Name == "" {
			_, _ = fmt.Fprintf(&sb2, "%s%s", arg.Type, nop)
		} else {
			_, _ = fmt.Fprintf(&sb2, "%s %s%s", arg.Name, arg.Type, nop)
		}
	}
	if len(m.OutputList) > 1 {
		_, _ = fmt.Fprintf(&sb, "(%s)", sb2.String())
	} else {
		_, _ = fmt.Fprintf(&sb, "%s", sb2.String())
	}
	var outNames strings.Builder
	var assertSet strings.Builder
	for i := 0; i < len(m.OutputList)-1; i++ {
		_, _ = fmt.Fprintf(&assertSet, "r%d,_ := reps[%d].(%s);", i, i, m.OutputList[i].Type)
		_, _ = fmt.Fprintf(&outNames, "r%d,", i)
	}
	if len(m.OutputList) == 1 {
		_, _ = fmt.Fprintf(&sb, "{_,err := %s.Call(\"%s\",%s);%sreturn %serr}", m.Receive.Name, m.ServiceName,
			inputNames.String(), assertSet.String(), outNames.String())
	} else if len(m.OutputList) > 1 {
		_, _ = fmt.Fprintf(&sb, "{reps,err := %s.Call(\"%s\",%s);%sreturn %serr}", m.Receive.Name, m.ServiceName,
			inputNames.String(), assertSet.String(), outNames.String())
	}
	return sb.String()
}

func (m *Method) FormatToASync() string {
	return ""
}

func (m *Method) FormatToRequests() string {
	return ""
}
