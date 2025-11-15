.PHONY: swagger_gen docker_build docker_push docker_deploy docker_compose_up docker_compose_down metrics_test

docker_build:
	docker build -t sso -t phenirain/fourthcoursefirstproject-sso:latest -f ./build/Dockerfile .

docker_push:
	docker push phenirain/fourthcoursefirstproject-sso:latest

docker_deploy: docker_build docker_push

docker_restart: docker_deploy docker_compose_down docker_compose_up

docker_compose_up:
	docker compose -f ./deployments/docker-compose.yaml up -d

docker_compose_down:
	docker compose -f ./deployments/docker-compose.yaml down

docker_compose_logs:
	docker compose -f ./deployments/docker-compose.yaml logs -f

swagger_gen:
	swag init -g cmd/sso/main.go -o docs --parseDependency --parseInternal

restart_with_config:
	docker cp config/config.yaml sso:/app/config/config.yaml
	docker restart sso

# Metrics commands
metrics_test:
	@echo "Testing metrics endpoint..."
	@curl -s http://localhost:8081/metrics | head -20

metrics_open:
	@echo "Opening monitoring interfaces..."
	@open http://localhost:8081/metrics
	@open http://localhost:9090
	@open http://localhost:3000