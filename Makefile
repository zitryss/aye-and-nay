.PHONY: setup compile test test-int test-ci dev-up dev-down prod-up prod-down

VM=192.168.99.106

setup:
	mkdir ./assets/tls/ && mkcert -cert-file ./assets/tls/cert.pem -key-file ./assets/tls/key.pem localhost ${VM} && chmod +r ./assets/tls/key.pem

compile:
	CGO_ENABLED=0 go build -ldflags='-s -w'

test:
	go test -v -race -cover ./...

test-int:
	go test -v -race -cover -tags=integration ./...

test-ci:
	go test -v -race -tags=integration -failfast -coverprofile=coverage.txt -covermode=atomic ./...

dev-up:
	docker-compose --file ./build/dev/docker-compose.yml up -d --build

dev-down:
	docker-compose --file ./build/dev/docker-compose.yml down --rmi all -v

prod-up:
	docker-compose --file ./build/prod/docker-compose.yml up -d --build

prod-down:
	docker-compose --file ./build/prod/docker-compose.yml down --rmi all -v
