.PHONY: git
git:
	git add .
	git commit -m"自动提交 git 代码"
	git push
tag:
	git push --tags
.PHONY: rpc
rpc:
	micro api  --handler=rpc  --namespace=go.micro.api --address=:8080
.PHONY: api
api:
	micro api  --handler=api  --namespace=go.micro.api --address=:8081
micro:
	micro api --enable_rpc=true

.PHONY: proto
proto:
	protoc -I . --micro_out=. --gogofaster_out=. proto/trade/trade.proto

.PHONY: docker
docker:
	docker build -f Dockerfile  -t vipspt .
.PHONY: run
run:
	go run main.go
test:
	go test main_test.go -test.v