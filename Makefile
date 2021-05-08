NAME := event-api-server

.PHONY: build
buid:
	go build ./cmd/$(NAME)

.PHONY: docker-push
docker-push:
	docker build -t asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/$(NAME):latest .
	docker push asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/$(NAME):latest

.PHONY: deploy
deploy:
	gcloud run deploy event-api-server \
		--platform=managed \
		--region=asia-east1 \
		--allow-unauthenticated \
		--image=asia.gcr.io/$(GOOGLE_CLOUD_PROJECT)/$(NAME):latest \
		--set-env-vars="SLACK_SIGNING_SECRET=$(SLACK_SIGNING_SECRET)" \
		--quiet
