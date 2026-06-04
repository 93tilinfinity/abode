# Common commands. Run `make check` before committing.
.PHONY: build test fmt vet lint run check

build: ; go build ./...
test:  ; go test ./...
fmt:   ; gofmt -l -w .
vet:   ; go vet ./...
lint:  ; golangci-lint run
run:   ; go run ./cmd/abode-daily
check: fmt vet build test
