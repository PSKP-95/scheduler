build:
	go build -tags hook_1

run: build
	./schedular.exe

migrateup:
	migrate -path db/migration -database "postgresql://postgres:password@172.29.149.60:5432/schedular?sslmode=disable" -verbose up

.PHONY: build migrateup run