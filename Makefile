build:
	@echo "Building..."
	@go build -o bin/http

run: build
	@echo "Starting application..."
	go run main.go httpd

test: 
	@echo "Running tests..."
	@go test -v ./...

# test-env:
# 	@echo "Sourcing envs:"
# 	@echo ${TEST_MAKE}ds