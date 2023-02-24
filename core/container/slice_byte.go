package container

type ByteSlice []byte

func (s *ByteSlice) Len() int { return len(*s) }

func (s *ByteSlice) Cap() int { return cap(*s) }

func (s *ByteSlice) Unique() {
	mark := make(map[byte]struct{}, len(*s))
	var count int
	for _, v := range *s {
		if _, ok := mark[v]; !ok {
			(*s)[count] = v
			mark[v] = struct{}{}
			count++
		}
	}
	*s = (*s)[:count]
}

func (s *ByteSlice) AppendSingle(v byte) {
	*s = append(*s, v)
}

func (s *ByteSlice) Append(v []byte) {
	*s = append(*s, v...)
}

func (s *ByteSlice) AppendS(vs ...byte) {
	*s = append(*s, vs...)
}

func (s *ByteSlice) Available() bool {
	return s != nil && len(*s) > 0
}

func (s *ByteSlice) Reset() {
	*s = (*s)[:0]
}
