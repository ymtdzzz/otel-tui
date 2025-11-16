.PHONY: help
help: ## Print help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-lint-tools: ## Install lint tools
	go install tool

DIR=./...
lint: ## Run static analysis
	go vet "$(DIR)"
	test -z "`gofmt -s -d .`"
	staticcheck "$(DIR)"
	prealloc -set_exit_status "$(DIR)"
	gosec "$(DIR)"

lint-exporter: ## Run static analysis for exporter
	$(MAKE) lint DIR="./tuiexporter/..."

.PHONY: test
test: ## run test ex.) make test OPT="-run TestXXX"
	TZ=UTC go test -v "$(DIR)" $(OPT)

test-coverage: ## Run test with coverage
	$(MAKE) test OPT="-coverprofile=coverage.out -coverpkg=./..."

test-coverage-report: ## Run test with coverage and generate report
	$(MAKE) test-coverage
	go tool cover -html=coverage.out

test-exporter: ## run test for exporter
	$(MAKE) test DIR="./tuiexporter/..."

test-coverage-exporter: ## Run test with coverage for exporter
	$(MAKE) test-coverage DIR="./tuiexporter/..."

test-coverage-report-exporter: ## Run test with coverage for exporter and generate report
	$(MAKE) test-coverage-report DIR="./tuiexporter/..."

update-screenshot: ## Update screenshot for docs
	@command -v vhs > /dev/null || (echo "vhs is needed. see: https://github.com/charmbracelet/vhs?tab=readme-ov-file#installation" && exit 1)
	go build -o otel-tui
	vhs screenshot.tape
	rm ./out.gif
	rm ./otel-tui
