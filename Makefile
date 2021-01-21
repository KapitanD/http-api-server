PROJECT?=github.com/KapitanD/http-api-server
APP?=apiserver
PORT?=8090
USER?=kapitand
CONTAINER_IMAGE?=docker.io/kapitand/${APP}

GOOS?=linux
GOARCH?=amd64

RELEASE?=0.0.2

.PHONY: build
build: 
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -v ./cmd/apiserver

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

container: build
	docker build -t $(CONTAINER_IMAGE):$(RELEASE) -f Dockerfile.dev .

run: container
	docker stop $(USER)/$(APP):$(RELEASE) || true && docker rm $(USER)/$(APP):$(RELEASE) || true
	docker run --name ${APP} -p ${PORT}:${PORT} --network=http-api-server_default --rm \
		-e "PORT=${PORT}" \
		$(USER)/$(APP):$(RELEASE)

push: container
	docker push $(CONTAINER_IMAGE):$(RELEASE)

minikube-up: push
	kubectl apply -f ./kubernetes/http-api-server/namespace.yaml
	kubectl apply -f ./kubernetes/http-api-server/configmap.yaml
	kubectl apply -f ./kubernetes/http-api-server/secret.yaml
	kubectl apply -f ./kubernetes/http-api-server/deployment.yaml
	kubectl apply -f ./kubernetes/http-api-server/service.yaml
	kubectl apply -f ./kubernetes/http-api-server/ingress.yaml

minikube-up-db:
	kubectl apply -f ./kubernetes/http-api-server/namespace.yaml
	kubectl apply -f ./kubernetes/postgresdb/config_postgres.yaml
	kubectl apply -f ./kubernetes/postgresdb/persvol_claim.yaml
	kubectl apply -f ./kubernetes/postgresdb/service_deployment.yaml

minikube-down:
	kubectl delete -f ./kubernetes/http-api-server/namespace.yaml