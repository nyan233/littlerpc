package main

type HelloServer1 int

func (s *HelloServer1) Hello() (string, error) {
	return "my is server 1", nil
}

type HelloServer2 string

func (s *HelloServer2) Init(str string) error {
	*s = HelloServer2(str)
	return nil
}

func (s *HelloServer2) Hello() (string, error) {
	return string(*s), nil
}
