#!/bin/bash

SERVICE_NAME=$1

if [ $SERVICE_NAME == "service-broker" ]; then
	gcloud run deploy ${SERVICE_NAME} \
		--platform=managed \
		--region=asia-east1 \
		--allow-unauthenticated \
		--image=asia.gcr.io/${GOOGLE_CLOUD_PROJECT}/${SERVICE_NAME}:latest \
		--set-env-vars="SLACK_SIGNING_SECRET=${SLACK_SIGNING_SECRET}" \
		--set-env-vars="SLACK_BOT_TOKEN=${SLACK_BOT_TOKEN}" \
		--quiet
elif [ $SERVICE_NAME == "example" ]; then
	gcloud run deploy ${SERVICE_NAME} \
		--platform=managed \
		--region=asia-east1 \
		--image=asia.gcr.io/${GOOGLE_CLOUD_PROJECT}/${SERVICE_NAME}:latest \
		--set-env-vars="SLACK_BOT_TOKEN=${SLACK_BOT_TOKEN}" \
		--quiet
fi