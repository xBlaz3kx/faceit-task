proto:
	protoc --go_out=./internal/grpc \
	--go_opt=paths=source_relative \
  	--proto_path=./proto/v1 \
	--experimental_allow_proto3_optional \
	--go-grpc_out=./internal/grpc \
	--go-grpc_opt=paths=source_relative \
    proto/v1/*.proto