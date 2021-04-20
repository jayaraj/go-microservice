FROM golang as builder

WORKDIR /go/src/microservice
COPY . .
RUN make setup
RUN CGO_ENABLED=0 go build -v

FROM alpine
LABEL maintainer="jayaraj.esvar@gmail.com"
WORKDIR /home
COPY --from=builder /go/src/microservice/go-microservice /home
COPY --from=builder /go/src/microservice/conf /home/conf
CMD [ "/home/go-microservice" ]
