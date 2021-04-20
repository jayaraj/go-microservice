generate:
	buf beta mod update
	buf lint
	buf generate
	mv temporary/proto/* proto/openapi
	rm -r temporary
	statik -m -f -src ./proto/openapi/

setup:
	go mod tidy
	go get \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
        google.golang.org/protobuf/cmd/protoc-gen-go \
	google.golang.org/grpc/cmd/protoc-gen-go-grpc \
        github.com/rakyll/statik

