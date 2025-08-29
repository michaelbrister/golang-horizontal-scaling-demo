# Simple Makefile for common Docker Compose tasks

PROJECT_NAME=golang-horizontal-scaling-poc

.PHONY: up down build logs ps

up:
	docker compose up --build --scale app=3

down:
	docker compose down -v

build:
	docker compose build --no-cache

logs:
	docker compose logs -f

ps:
	docker compose ps

scale:
	docker compose up -d --scale app=$(N)
