build:
	go build -tags hook_1,hook_2,hook_3

run: build
	./scheduler.exe

lint:
	golangci-lint run

test:
	go test -v -coverprofile cover.out ./...

coverage:
	go tool cover -html=cover.out

mock:
	mockgen -destination db/mock/store.go -package mockdb -source ./db/sqlc/store.go Store

migrateup:
	migrate -path db/migration -database "postgresql://postgres:password@172.29.149.60:5432/scheduler?sslmode=disable" -verbose up

.PHONY: build migrateup run test lint