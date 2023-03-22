package main

type HelloServer1 struct{}

func (s *HelloServer1) Hello() (string, error) {
	return "my is server 1", nil
}

type HelloServer2 struct {
	S string
}

func (s *HelloServer2) Init(str string) error {
	s.S = str
	return nil
}

func (s *HelloServer2) Hello() (string, error) {
	return s.S, nil
}
