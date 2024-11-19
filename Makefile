proto:
	protoc --go_out=./internal/grpc \
	--go_opt=paths=source_relative \
  	--proto_path=./proto/v1 \
	--experimental_allow_proto3_optional \
	--go-grpc_out=./internal/grpc \
	--go-grpc_opt=paths=source_relative \
    proto/v1/*.proto

create-keyfile:
	@if [ ! -f "./keyfile" ]; then \
        openssl rand -base64 756 > ./keyfile; \
        sudo chmod 400 ./keyfile; \
        sudo chown 999:999 ./keyfile; \
        echo "Created keyfile"; \
	else \
		echo "Keyfile already exists"; \
    fi
