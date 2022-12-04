package main

import (
	"embed"
	"fmt"
)

var (
	//go:embed template
	templateFs embed.FS

	BeforeCodeTemplate string
)

func init() {
	bfData, err := templateFs.ReadFile("template/beforecode.gohtml")
	if err != nil {
		panic(err)
	}
	BeforeCodeTemplate = fmt.Sprintf(string(bfData), "\n", "\n")
}
