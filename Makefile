generate:
	go mod tidy 
	protoc \
        -I proto \
        -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway \
        -I ${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
        --go_out=plugins=grpc,paths=source_relative:proto \
        --grpc-gateway_out=./proto \
        --openapiv2_out=proto/openapi/ \
        proto/*.proto

	mv proto/go-microservice/proto/* proto
	
	rm -r proto/go-microservice

	statik -m -f -src ./proto/openapi/

setup:
	go install \
        github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
        github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
        github.com/golang/protobuf/protoc-gen-go \
        github.com/rakyll/statik
