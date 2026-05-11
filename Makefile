.PHONY: up down test test-go test-python test-ts lint build clean

up:
	docker compose up --build -d

down:
	docker compose down

build:
	docker compose build

test: test-go test-python test-ts

test-go:
	cd gateway && go test -v ./...

test-python:
	cd monitor && pip install -q -r requirements.txt -r requirements-dev.txt && python -m pytest tests/ -v

test-ts:
	cd dashboard && npm install && npm test

lint: lint-go lint-python lint-ts

lint-go:
	cd gateway && go vet ./...

lint-python:
	cd monitor && pip install -q flake8 && flake8 app.py --max-line-length=120

lint-ts:
	cd dashboard && npm install && npx eslint src/

logs:
	docker compose logs -f

status:
	@curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || echo "Gateway not running"
	@curl -s http://localhost:8081/health | python3 -m json.tool 2>/dev/null || echo "Monitor not running"
	@curl -s http://localhost:8082/health | python3 -m json.tool 2>/dev/null || echo "Dashboard not running"

clean:
	docker compose down --rmi local --volumes
	rm -rf dashboard/node_modules dashboard/dist
	rm -rf monitor/__pycache__ monitor/tests/__pycache__
