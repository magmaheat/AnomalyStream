COMPOSE_FILE = ./config/docker-compose.yml
SERVICE_NAME = db

.PHONY: all clean up down restart

all: clean install

build:
	mkdir -p "../build"

install: build
	go build -o ../build/server server/main.go
	go build -o ../build/client client/main.go

run: up
	./../build/server

up:
	docker-compose -f $(COMPOSE_FILE) up -d

down:
	docker-compose -f $(COMPOSE_FILE) down

restart: down up

logs:
	docker-compose -f $(COMPOSE_FILE) logs -f $(SERVICE_NAME)

clean: down
	rm -rf ../build
