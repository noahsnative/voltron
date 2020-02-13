BINARY := voltron-injector

PLATFORMS := linux darwin windows
os = $(word 1, $@)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
		mkdir -p bin
		GO111MODULE=on GOOS=$(os) GOARCH=amd64 go build -o ./bin/$(BINARY)-$(os)-amd64 -v ./cmd/voltron-injector/main.go
