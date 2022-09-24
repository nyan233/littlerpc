package test

func (p *Test) Proxy2(s1 string, s2 int) (r1 *Test, r2, r3 int64, err error) { return nil, 0, 0, nil }

func (p *Test) MapCallTest(m1, m2 map[string]map[string]byte) (*Test, map[string]byte, error) {
	return nil, nil, nil
}

func (p *Test) CallSlice(s1 []*Test, s2 []map[string]int) (bool, error) { return false, nil }
