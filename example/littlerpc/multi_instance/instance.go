package main

type HelloServer1 int

func (s *HelloServer1) Hello() string {
	return "my is server 1"
}

type HelloServer2 string

func (s *HelloServer2) Init(str string) {
	*s = HelloServer2(str)
}

func (s *HelloServer2) Hello() string {
	return string(*s)
}