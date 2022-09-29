# 包功能的划分
```text
pkg/* --> 所有模块都可能依赖的包, 修改其中的公共设置可能会影响到所有使用到它们的模块, 包括client&server
pkg/common/* --> 只有client&server会使用, 使用其中的一些公共设置只会影响这两个
```