generate_grpc_code:
	protoc \
	--go_out=./internal \
	--go_opt=paths=source_relative \
	--go-grpc_out=./internal \
	--go-grpc_opt=paths=source_relative \
	proto/msg.proto

build_worker:generate_grpc_code
	mkdir -p bin
	go build -o bin/worker cmd/worker/main.go 
build_coordinator:generate_grpc_code
	mkdir -p bin
	go build -o bin/coordinator cmd/coordinator/main.go

build_worker_amd64:
	mkdir -p bin
	GOARCH=amd64 GOOS=linux go build -o bin/worker-linux-amd64 cmd/worker/main.go
	tar -czvf bin/worker-linux-amd64.tar.gz bin/worker-linux-amd64
build_worker_arm64:
	mkdir -p bin
	GOARCH=arm64 GOOS=linux go build -o bin/worker-linux-arm64 cmd/worker/main.go
	tar -czvf bin/worker-linux-arm64.tar.gz bin/worker-linux-arm64

build_coordinator_amd64:
	mkdir -p bin
	GOARCH=amd64 GOOS=linux go build -o bin/coordinator-linux-amd64 cmd/coordinator/main.go
	tar -czvf bin/coordinator-linux-amd64.tar.gz bin/coordinator-linux-amd64
build_coordinator_arm64:
	mkdir -p bin
	GOARCH=arm64 GOOS=linux go build -o bin/coordinator-linux-arm64 cmd/coordinator/main.go
	tar -czvf bin/coordinator-linux-arm64.tar.gz bin/coordinator-linux-arm64