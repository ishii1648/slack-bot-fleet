## service-broker

.PHONY: build-service-broker
buid-service-broker:
	go build ./cmd/service-broker

.PHONY: docker-push-service-broker
docker-push-service-broker:
	docker build -t asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/service-broker:latest . -f cmd/service-broker/Dockerfile
	docker push asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/service-broker:latest

.PHONY: deploy-service-broker
deploy-service-broker: docker-push-service-broker
	bash ./deploy.sh service-broker

## chatbot

.PHONY: build-chatbot
buid-chatbot:
	go build ./cmd/chatbot

.PHONY: docker-push-chatbot
docker-push-chatbot:
	docker build -t asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/chatbot:latest . -f cmd/chatbot/Dockerfile
	docker push asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/chatbot:latest

.PHONY: deploy-chatbot
deploy-chatbot: docker-push-chatbot
	bash ./deploy.sh chatbot

## gen proto

.PHONY: go-proto
go-proto:
	protoc -I api/services/chatbot/ --go_out=plugins=grpc,paths=source_relative:api/services/chatbot/ api/services/chatbot/*.proto
