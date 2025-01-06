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