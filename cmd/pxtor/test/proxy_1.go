package test

type Test struct{}

func (p *Test) Foo(s1 string) (int, error) {
	return 1 << 20, nil
}

func (p *Test) Bar(s1 string) (int, error) {
	return 1 << 30, nil
}

func (p *Test) NoReturnValue(i int) error {
	return nil
}
