BINARY_NAME := gocrawl
GO := go
DOCKER := docker
IMAGE_NAME := gocrawl

build:
	$(GO) build -o $(BINARY_NAME) ./cmd/gocrawl

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

lint:
	golangi-lint run ./...

run:
	$(GO) run ./cmd/gocrawl $(ARGS)

docker-build:
	$(DOCKER) build -t $(IMAGE_NAME) .

docker-run:
	$(DOCKER) run --rm -it -v $(PWD):/data $(IMAGE_NAME) $(ARGS)

clean:
	rm -f $(BINARY_NAME)
	$(DOCKER) rmi $(IMAGE_NAME) 2>/dev/null || true

help:
	@echo "Available commands:"
	@echo "  make build       			- Build binary"
	@echo "  make test        			- Run tests"
	@echo "  make test-race   			- Run tests with race detector"
	@echo "  make lint        			- Run linter"
	@echo "  make run ARGS='...' 		- Run application"
	@echo "  make docker-build 			- Build Docker image"
	@echo "  make docker-run ARGS='...' - Run Docker container"
	@echo "  make clean       			- Cleanup"
	@echo "  make help      			- Show help"