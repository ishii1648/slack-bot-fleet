FROM golang:1.16.9 as builder

WORKDIR /go/src/

COPY . .
RUN go mod download

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
RUN go build -o ./bin/google-cloud-bot ./cmd/google-cloud-bot
FROM alpine:3.12 as runner

COPY --from=builder /go/src/bin/google-cloud-bot ./google-cloud-bot
COPY --from=builder /go/src/*.yml ./

ENTRYPOINT ["./google-cloud-bot"]