.PHONY: swagger_gen docker_build docker_push docker_deploy docker_compose_up docker_compose_down

docker_build:
	docker build -t sso -t phenirain/fourthcoursefirstproject-sso:latest -f ./build/Dockerfile .

docker_push:
	docker push phenirain/fourthcoursefirstproject-sso:latest

docker_deploy: docker_build docker_push

docker_compose_up:
	docker compose -f ./deployments/docker-compose.yaml up -d

docker_compose_down:
	docker compose -f ./deployments/docker-compose.yaml down

swagger_gen:
	swag init -g cmd/sso/main.go -o docs --parseDependency --parseInternal