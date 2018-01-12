FROM golang:alpine AS builder
RUN apk --no-cache add \
    build-base \
    bash \
    git
ADD . /go/src/github.com/thingful/big-iot-gateway-test
WORKDIR /go/src/github.com/thingful/big-iot-gateway-test
RUN echo 'Building binary with: ' && go version
RUN make build

FROM alpine
RUN apk --no-cache add \
    ca-certificates
WORKDIR /app
EXPOSE 8080
COPY --from=builder /go/src/github.com/thingful/big-iot-gateway-test/build/big-iot-gateway /app/
RUN /app/big-iot-gateway