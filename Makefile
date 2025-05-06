help: ## Show this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+%?:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install binary
	@go install .

serve: ## Run server with dev config
	@go run . serve

test: ## Run server tests
	@go test -cover ./...

lint: ## Run full linting rules
	@go vet ./...
	@golangci-lint run
	@staticcheck ./...

clean: ## Remove all previous build artifacts
	@rm -f $(which dicom-microservice-api)
	@rm -rf ./tmp
