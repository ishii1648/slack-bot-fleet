.PHONY: run
run:
	go run ./cmd/google-cloud-bot -debug

.PHONY: test
	go test -v ./...