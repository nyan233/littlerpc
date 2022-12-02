package main

import (
	"fmt"
	"io/ioutil"
)

var (
	BeforeCodeTemplate string
)

func init() {
	bfData, err := ioutil.ReadFile("./template/beforecode.gohtml")
	if err != nil {
		panic(err)
	}
	BeforeCodeTemplate = fmt.Sprintf(string(bfData), "\n", "\n")
}
