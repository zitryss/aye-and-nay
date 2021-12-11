.PHONY: gen compile test-unit test-int test-unit-ci test-int-ci dev-up dev-down prod-loadtest prod-up prod-down embed-loadtest embed-up embed-down

gen:
	go install github.com/mailru/easyjson/easyjson@latest
	go install golang.org/x/tools/cmd/stringer@latest
	go generate ./...

compile: gen
	CGO_ENABLED=0 go build -ldflags="-s -w"

test-unit: gen
	go test -v -race -shuffle=on -count=1 -short -tags=unit -cover ./...

test-int: gen
	go test -v -race -shuffle=on -count=1 -short -tags=integration -cover ./...

test-unit-ci: gen
	go test -v -race -shuffle=on -count=1 -tags=unit -failfast -coverprofile=coverage.txt -covermode=atomic ./...

test-int-ci: gen
	go test -v -race -shuffle=on -count=1 -tags=integration -failfast -coverprofile=coverage.txt -covermode=atomic ./...

dev-up:
	docker compose --file ./build/docker-compose-dev.yml up -d --build

dev-down:
	docker compose --file ./build/docker-compose-dev.yml down --rmi all -v

prod-loadtest:
	go run ./cmd/loadtest/main.go -verbose=false

prod-up:
	docker compose --file ./build/docker-compose-prod.yml up -d --build

prod-down:
	docker compose --file ./build/docker-compose-prod.yml down --rmi all -v

embed-loadtest:
	go run ./cmd/loadtest/main.go -verbose=false -api-address "http://localhost:8001" -minio-address ""

embed-up:
	docker compose --file ./build/docker-compose-embed.yml up -d --build

embed-down:
	docker compose --file ./build/docker-compose-embed.yml down --rmi all -v
