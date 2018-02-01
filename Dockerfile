FROM golang:alpine AS builder
RUN apk --no-cache add \
    build-base \
    bash \
    git
ADD . /go/src/github.com/thingful/big-iot-gateway
WORKDIR /go/src/github.com/thingful/big-iot-gateway
RUN echo 'Building binary with: ' && go version
RUN make build

FROM alpine
RUN apk --no-cache add \
    ca-certificates
WORKDIR /app
EXPOSE 8080
COPY --from=builder /go/src/github.com/thingful/big-iot-gateway/build/big-iot-gateway /app/
ADD ./config.yaml /app/
ADD ./offers.json /app/
CMD /app/big-iot-gateway start --config config.yaml --providerID=Thingful_test-Thingful_test_provider --providerSecret=ZoGeUUfnShyDZSasrlF-Rw== --pipeAccessToken=b9a99462-a031-41aa-aa21-8aacd738085a