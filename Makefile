proto:
	protoc --go_out=./pkg/proto/v1 \
	--go_opt=paths=source_relative \
  	--proto_path=./pkg/proto/v1 \
	--experimental_allow_proto3_optional \
	--go-grpc_out=./pkg/proto/v1 \
	--go-grpc_opt=paths=source_relative \
    pkg/proto/v1/*.proto