.PHONY: build run test vet fmt local-up local-down local-logs migrate local-reset seed

BINARY := bin/vidhya-service

## Compile the service binary into ./bin.
build:
	go build -o $(BINARY) ./src

## Run the service directly with go run (reads ./.env if you `source .env` first).
run:
	go run ./src

## go vet ./...
vet:
	go vet ./...

## gofmt -l on the whole tree (fails if anything is unformatted).
fmt:
	gofmt -l .

## go test ./...
test:
	go test ./...

## Start local Postgres + Redis (Kafka stays down; see docker-compose --profile kafka).
local-up:
	docker compose -f devops/local/docker-compose.yml --env-file devops/local/.env up -d postgres redis

## Stop and remove local containers (data volumes are preserved).
local-down:
	docker compose -f devops/local/docker-compose.yml --env-file devops/local/.env down

## Tail logs from all local containers.
local-logs:
	docker compose -f devops/local/docker-compose.yml --env-file devops/local/.env logs -f

## Apply SQL changelogs (sql/changelogs) to the local database via Liquibase.
migrate:
	docker compose -f devops/local/docker-compose.yml --env-file devops/local/.env up --build liquibase

## Danger: wipes the local Postgres volume so the next `local-up` starts fresh.
local-reset:
	docker compose -f devops/local/docker-compose.yml --env-file devops/local/.env down -v
