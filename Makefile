.PHONY: gen compile test test-int test-ci dev-up dev-down prod-up prod-down embed-up embed-down loadtest

gen:
	go install github.com/mailru/easyjson/easyjson@latest
	go generate ./...

compile: gen
	CGO_ENABLED=0 go build -ldflags="-s -w"

test: gen
	go test -v -race -count=1 -short -cover ./...

test-int: gen
	go test -v -race -count=1 -short -cover -tags=integration ./...

test-ci: gen
	go test -v -race -count=1 -tags=integration -failfast -coverprofile=coverage.txt -covermode=atomic ./...

dev-up:
	docker-compose --file ./build/docker-compose-dev.yml up -d --build

dev-down:
	docker-compose --file ./build/docker-compose-dev.yml down --rmi all -v

prod-loadtest:
	go run ./cmd/loadtest/main.go -verbose=false

prod-up:
	docker-compose --file ./build/docker-compose-prod.yml up -d --build

prod-down:
	docker-compose --file ./build/docker-compose-prod.yml down --rmi all -v

embed-loadtest:
	go run ./cmd/loadtest/main.go -verbose=false -api-address "http://localhost:8001" -minio-address ""

embed-up:
	docker-compose --file ./build/docker-compose-embed.yml up -d --build

embed-down:
	docker-compose --file ./build/docker-compose-embed.yml down --rmi all -v
