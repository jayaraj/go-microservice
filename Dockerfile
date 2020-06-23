FROM golang as builder

WORKDIR /go/src/microservice
COPY . .
RUN apt-get update
RUN apt-get install -y unzip
ENV PROTOBUF_VERSION 3.12.3
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOBUF_VERSION/protoc-$PROTOBUF_VERSION-linux-x86_64.zip
RUN unzip protoc-$PROTOBUF_VERSION-linux-x86_64.zip -d /usr/local/
RUN rm -rf protoc-$PROTOBUF_VERSION-linux-x86_64.zip
RUN GO111MODULE=off go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
RUN make setup
RUN make generate
RUN CGO_ENABLED=0 go build -v

FROM alpine
LABEL maintainer="jayaraj.esvar@gmail.com"
WORKDIR /home
COPY --from=builder /go/src/microservice/go-microservice /home
COPY --from=builder /go/src/microservice/conf /home/conf
CMD [ "/home/go-microservice" ]
