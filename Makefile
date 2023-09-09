build:
	go build -tags hook_1,hook_2,hook_3

run: build
	./scheduler.exe

lint:
	golangci-lint run

test:
	go test -v -cover ./...

migrateup:
	migrate -path db/migration -database "postgresql://postgres:password@172.29.149.60:5432/scheduler?sslmode=disable" -verbose up

.PHONY: build migrateup run test lint