dev:
	@go mod tidy
	@go run *.go

init:
	@docker-compose up -d