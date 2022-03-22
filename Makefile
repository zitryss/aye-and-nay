.PHONY: gen compile compile-health test-unit test-int test-unit-ci test-int-ci dev-up dev-down prod-loadtest prod-up prod-down embed-loadtest embed-up embed-down

gen:
	go install github.com/mailru/easyjson/easyjson@latest
	go install golang.org/x/tools/cmd/stringer@latest
	go generate ./...

compile: gen
	CGO_ENABLED=0 go build -ldflags="-s -w"

compile-health:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o healthcheck ./cmd/healthcheck/main.go

test-unit: gen
	go test -v -race -shuffle=on -count=2 -short -cover ./... -args -unit

test-int: gen
	go test -v -race -shuffle=on -count=2 -short -cover ./... -args -int

test-unit-ci: gen
	go test -v -race -shuffle=on -count=2 -failfast -coverprofile=coverage.txt -covermode=atomic ./... -args -unit -ci

test-int-ci: gen
	go test -v -race -shuffle=on -count=2 -failfast -coverprofile=coverage.txt -covermode=atomic ./... -args -int -ci

dev-up:
	docker compose --file ./build/docker-compose-dev.yml up -d --build

dev-down:
	docker compose --file ./build/docker-compose-dev.yml down --rmi all -v

prod-loadtest:
	go run ./cmd/loadtest/* -verbose=false

prod-up:
	docker compose --file ./build/docker-compose-prod.yml up -d --build

prod-down:
	docker compose --file ./build/docker-compose-prod.yml down --rmi all -v

embed-loadtest:
	go run ./cmd/loadtest/* -verbose=false -api-address "http://localhost:8001" -minio-address "http://localhost:9000" -html-address ""

embed-up:
	docker compose --file ./build/docker-compose-embed.yml up -d --build

embed-down:
	docker compose --file ./build/docker-compose-embed.yml down --rmi all -v
