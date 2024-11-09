run:
	@go run cmd/api/main.go

migration:
	@go run cmd/migration/main.go

test:
	@go test -v ./...

build:
	@go build -o bin/api cmd/api/main.go