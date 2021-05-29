.PHONY: build
build: build-service-broker

## service-broker

.PHONY: build-service-broker
build-service-broker:
	go build  -o ./bin/service-broker ./cmd/service-broker

.PHONY: docker-push-service-broker
docker-push-service-broker:
	docker build -t asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/service-broker:latest . -f cmd/service-broker/Dockerfile
	docker push asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/service-broker:latest

.PHONY: deploy-service-broker
deploy-service-broker: docker-push-service-broker
	bash ./deploy.sh service-broker

## example

.PHONY: build-example
build-example:
	go build -o ./bin/example ./cmd/example

.PHONY: docker-push-example
docker-push-example:
	docker build -t asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/example:latest . -f cmd/example/Dockerfile
	docker push asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/example:latest

.PHONY: deploy-example
deploy-example: docker-push-example
	bash ./deploy.sh example

## test
.PHONY: test
test:
	go test -v ./service/...

## gen proto

.PHONY: go-proto
go-proto:
	protoc -I api/services/example/ \
		--go_out=api/services/example/ --go_opt=paths=source_relative \
		--go-grpc_out=api/services/example/ --go-grpc_opt=paths=source_relative \
		api/services/example/*.proto

## run
.PHONY: run-service-broker
run-service-broker:
	go run ./cmd/service-broker/main.go -disable-auth

.PHONY: run example
run-example:
	go run ./cmd/example/main.go