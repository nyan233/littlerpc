package test

import "errors"

type Test struct {}

func (p *Test) Foo(s1 string) int {
	return 1 << 20
}

func (p *Test) Bar(s1 string) int {
	return 1 << 30
}

func (p *Test) ErrHandler(s1 string) (err error) {
	return errors.New(s1)
}