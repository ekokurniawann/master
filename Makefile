.PHONY: run swag swag-init clean redis-up redis-down

run:
	go run main.go

swag:
	swag init -g main.go --parseDependency --parseInternal

swag-init: swag run

redis-up:
	podman compose -f redis-service.yaml up -d

redis-down:
	podman compose -f redis-service.yaml down

clean:
	go clean
