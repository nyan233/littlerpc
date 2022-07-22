package container

type Slice[V comparable] []V

func (s *Slice[V]) Len() int { return len(*s) }

func (s *Slice[V]) Cap() int { return cap(*s) }

func (s *Slice[V]) Unique() {
	mark := make(map[V]struct{}, len(*s))
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

func (s *Slice[V]) Append(v []V) {
	*s = append(*s, v...)
}

func (s *Slice[V]) AppendS(vs ...V) {
	*s = append(*s, vs...)
}
