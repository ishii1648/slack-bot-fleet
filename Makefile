.PHONY: build
build: build-service-broker build-example

## service-broker

.PHONY: build-service-broker
build-service-broker:
	go build  -o ./bin/service-broker ./cmd/service-broker

.PHONY: docker-push-service-broker
docker-push-service-broker:
	docker build -t asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/service-broker:latest . -f Dockerfile-service-broker
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
	docker build -t asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/example:latest . -f Dockerfile-example
	docker push asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/example:latest

.PHONY: deploy-example
deploy-example: docker-push-example
	bash ./deploy.sh example

## test
.PHONY: test
test:
	bash ./test.sh

.PHONY: test-coverage
test-coverage:
	bash ./test.sh -coverage

## gen proto

.PHONY: go-proto
go-proto:
	protoc -I proto/reaction-added-event/ \
		--go_out=proto/reaction-added-event/ --go_opt=paths=source_relative \
		--go-grpc_out=proto/reaction-added-event/ --go-grpc_opt=paths=source_relative \
		proto/reaction-added-event/*.proto

## run
.PHONY: run
run: run-service-broker

.PHONY: run-service-broker
run-service-broker:
	go run ./cmd/service-broker/main.go -disable-auth -debug

.PHONY: run example
run-example:
	go run ./cmd/example/main.go -debug