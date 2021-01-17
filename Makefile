.PHONY: build
build: 
	go build -v ./cmd/apiserver

.PHONY: testify
test:
	go test -v -race -timeout 30s ./...

.DEFAULT_GOAL := build

.PHONY: migrations-dev
migrations-dev:
	migrate -path migrations -database "postgres://localhost:12345/restapi_dev?sslmode=disable&&user=postgres&&password=example" $(METHOD)

.PHONY: migrations-test
migrations-test:
	migrate -path migrations -database "postgres://localhost:12345/restapi_test?sslmode=disable&&user=postgres&&password=example" $(METHOD)