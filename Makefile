.PHONY: swagger_gen docker_build docker_push docker_deploy docker_compose_up docker_compose_down metrics_test alerts_test

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

# Alerting commands
alerts_open:
	@echo "Opening alerting interfaces..."
	@open http://localhost:9090/alerts
	@open http://localhost:9093

alerts_test_firing:
	@echo "🔥 Triggering HighHTTPErrorRate alert (generating 401 errors)..."
	@for i in {1..20}; do curl -s http://localhost:8081/admin/clients > /dev/null; sleep 0.5; done
	@echo "✅ Done! Check Telegram for FIRING alert in ~30 seconds"

alerts_test_resolved:
	@echo "✅ Resolving alert (generating successful requests)..."
	@curl -s -X POST http://localhost:8081/auth/signUp -H "Content-Type: application/json" -d '{"login":"alerttest","password":"test123"}' > /dev/null
	@for i in {1..20}; do curl -s -X POST http://localhost:8081/auth/logIn -H "Content-Type: application/json" -d '{"login":"alerttest","password":"test123"}' > /dev/null; sleep 0.5; done
	@echo "✅ Done! Check Telegram for RESOLVED alert in ~1 minute"

alerts_logs:
	@echo "📋 Webhook logs:"
	@docker logs --tail 50 sso-webhook

alerts_status:
	@echo "📊 Alert status:"
	@curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | {state: .state, name: .labels.alertname}'