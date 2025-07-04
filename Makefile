ifneq (,$(wildcard .env))
  include .env
  export
endif

PROTO_DIR=pkg/proto
OUT_DIR=pkg/proto/gen/go/

MIGRATE ?= $(HOME)/go/bin/migrate
DATABASE_URL ?= "$(POSTGRES_DB)://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_NAME_DB)?sslmode=disable"
MIGRATIONS_DIR ?= ./migrations
.PHONY: up down

up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database $(DATABASE_URL) up

down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database $(DATABASE_URL) down

auth_db:
	docker compose up --build -d db

gen:
	protoc --go_out=$(OUT_DIR) --go-grpc_out=$(OUT_DIR) $(PROTO_DIR)/movie.proto

swag:
	swag init --parseDependency --parseInternal --generalInfo internal/delivery/http/server/docs/docs.go --output docs
