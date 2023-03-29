build:
	@go build -o bin/redis.a *.go

run: build
	@./bin/redis.a

