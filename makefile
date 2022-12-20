
cover-test:
	#go test -coverprofile=coverage.txt -covermode=atomic -v ./...
	# all cover
	go test -coverprofile=all.cover.out -covermode=atomic -v ./...
	# client&server&common cover
	go test -coverprofile=impl.cover.out -coverpkg="./core/client,./core/server,./core/common" -covermode=atomic -v -run=Test* ./test
	# 合并测试覆盖率
	sh merge_cover.sh

.PHONY:test
test:
	go test -v ./...

race-test:
	go test -race -v ./...

test-fuzz:
	go test -fuzz -v ./...

test-bench:
	go test -bench -v ./...

test-example:
