package test

func (p *Test) Proxy2(s1 string, s2 int) (r1 *Test, r2, r3 int64) { return nil, 0, 0 }

func (p *Test) MapCallTest(m1, m2 map[string]map[string]byte) (*Test, map[string]byte) {
	return nil, nil
}

func (p *Test) CallSlice(s1 []*Test, s2 []map[string]int) bool { return false }
