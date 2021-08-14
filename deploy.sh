#!/bin/bash

SERVICE_NAME=$1

if [ $SERVICE_NAME == "service-broker" ]; then
	gcloud run deploy ${SERVICE_NAME} \
		--platform=managed \
		--region=asia-northeast1 \
		--image=asia.gcr.io/${GOOGLE_CLOUD_PROJECT}/${SERVICE_NAME}:latest \
		--quiet
elif [ $SERVICE_NAME == "example" ]; then
	gcloud run deploy ${SERVICE_NAME} \
		--platform=managed \
		--region=asia-northeast1 \
		--image=asia.gcr.io/${GOOGLE_CLOUD_PROJECT}/${SERVICE_NAME}:latest \
		--quiet
fi