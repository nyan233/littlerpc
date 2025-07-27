COMMON_PKG = "./core/common/msgparser,./core/common/msgwriter,./core/common/transport,./core/common/context"
CORE_PKG = "./core/client,./core/server,./core/protocol,./core/utils"
init:



cover-test:init
	#go test -coverprofile=coverage.txt -covermode=atomic -v ./...
	# all cover
	go test -coverprofile="all.cover.out" -covermode=atomic -v ./...
	# client&server&common cover
	go test -coverprofile="impl.cover.out" -coverpkg=$(CORE_PKG) + "," + $(COMMON_PKG) -covermode=atomic -v -run=Test* ./test
	# 合并测试覆盖率
	sh merge_cover.sh

.PHONY:test
test:init
	go test -v ./...

race-test:init
	go test -race -v ./...

test-fuzz:init
	go test -fuzz -v ./...

test-bench:init
	go test -bench -v ./...

test-example:
