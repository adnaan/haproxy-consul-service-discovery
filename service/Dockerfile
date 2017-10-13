FROM golang:1.9.1 as builder
WORKDIR /go/src/github.com/adnaan/haproxy-consul-service-discovery/service
ADD ./cmd  ./cmd
RUN cd ./cmd && go get -v
RUN cd ./cmd && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sampleservice .

FROM alpine:latest
EXPOSE 3344
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/adnaan/haproxy-consul-service-discovery/service/cmd/sampleservice .
CMD ["./sampleservice"]