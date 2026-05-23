.PHONY: run swag swag-init clean

run:
	go run main.go

swag:
	swag init -g main.go --parseDependency --parseInternal

swag-init: swag run

clean:
	go clean