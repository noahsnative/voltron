BINARY := injector

PLATFORMS := linux darwin windows
os = $(word 1, $@)

# Builds for one of the specified platforms (linux, darwin, windows)
.PHONY: $(PLATFORMS)
$(PLATFORMS):
		mkdir -p bin
		GOOS=$(os) GOARCH=amd64 go build -o ./bin/$(BINARY)-$(os)-amd64 ./cmd/injector/main.go

# Runs all of the tests
.PHONY: test
test:
		go test -v ./...

# Builds a docker image
.PHONY: image
image:
		docker build -t voltron/injector:latest .
