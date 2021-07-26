FROM golang:1.15.8 as builder

WORKDIR /go/src/

COPY . .
RUN go mod download

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
RUN go build -o ./bin/service-broker ./cmd/service-broker

FROM alpine:3.12 as runner

COPY --from=builder /go/src/bin/service-broker ./service-broker
COPY --from=builder /go/src/template/*.yml ./template/

ENTRYPOINT ["./service-broker"]