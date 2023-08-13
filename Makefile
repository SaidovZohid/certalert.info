# use .env in Makefile
-include .env
# use .SILENT: to remove the printing of the commands when calling alias in Makefile
.SILENT:
# give get current direction(folder)
CURRENT_DIR=$(shell pwd)
# variable to make postgresql url
DB_URL=postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable

# go mod
tidy:
	@go mod tidy
	@go mod vendor

# this builds for server. before building golang project export ENV
# export GOOS=linux
# export GOARCH=amd64
build:
	GOOS=linux GOARCH=amd64 go build -o ./bin/main ./cmd/main.go

# database migrations
# 
# create migration name=users
create-migration:
	@migrate create -ext sql -dir migrations -seq $(name)
#
# up all migrations
migrateup:
	@migrate -path migrations -database "$(DB_URL)" -verbose up
#
# up migration last one
migrateup1:
	@migrate -path migrations -database "$(DB_URL)" -verbose up 1
#
# down migrations all
migratedown:
	@migrate -path migrations -database "$(DB_URL)" -verbose down
#
# down the migration last
migratedown1:
	@migrate -path migrations -database "$(DB_URL)" -verbose down 1
#
# syntax 
lint:
	@golangci-lint run ./...
#
# testing
#
# clean test cache
cache:
	@go clean -testcache
#
# run test
test:
	@go test -v -cover ./...

push:
	docker build --platform linux/amd64 -t zohiddev/certalert:latest .
	docker image push zohiddev/certalert:latest