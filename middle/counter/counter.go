package counter

// ServerCounter 计数器/统计器
// 实现它的实例必须是线程安全的，因为该接口会被不同的goroutine调用
type ServerCounter interface {
	CountTimestamp(rt int64, callT int64)
	CountRawBodyLength(call int64, ret int64)
	CountEncodingBodyLength(call int64, ret int64)
}
