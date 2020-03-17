BINARY := injector

PLATFORMS := linux darwin windows
os = $(word 1, $@)

# Install tools (linter, etc.)
.PHONY: install-tools
install-tools:
		wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.23.7

# Builds for one of the specified platforms (linux, darwin, windows)
.PHONY: $(PLATFORMS)
$(PLATFORMS):
		mkdir -p bin
		GOOS=$(os) GOARCH=amd64 go build -o ./bin/$(BINARY)-$(os)-amd64 ./cmd/injector/main.go

# Runs static code analysis. Make sure you run install-tools before first use
.PHONY: lint
lint:
		./bin/golangci-lint run ./...

# Runs unit tests
.PHONY: tests-unit
tests-unit:
		go test -v ./...


# Runs end to end tests
.PHONY: tests-e2e
tests-e2e:
		./test/e2e/run.sh

# Builds a docker image
.PHONY: docker
docker:
		docker build -t voltron/injector:latest -f ./deploy/injector/Dockerfile .
