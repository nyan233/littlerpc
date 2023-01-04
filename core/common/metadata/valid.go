package metadata

func (opt *ProcessOption) Valid() bool {
	if opt.SyncCall && opt.UseRawGoroutine {
		return false
	}
	return true
}
