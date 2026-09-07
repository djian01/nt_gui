.PHONY: setup frontend build run test mac-app

# Match the minimum used by Wails' linker when compiling native objects.
ifeq ($(shell go env GOOS),darwin)
export CGO_CFLAGS := $(CGO_CFLAGS) -mmacosx-version-min=11.0
export CGO_LDFLAGS := $(CGO_LDFLAGS) -mmacosx-version-min=11.0
endif

setup:
	npm --prefix frontend ci
	go mod download

frontend:
	npm --prefix frontend run build

build: frontend
	go build -tags production -o bin/nt-http .

run: build
	./bin/nt-http

mac-app: build
	mkdir -p "bin/NT HTTP Prototype.app/Contents/MacOS"
	cp bin/nt-http "bin/NT HTTP Prototype.app/Contents/MacOS/nt-http"
	cp build/darwin/Info.plist "bin/NT HTTP Prototype.app/Contents/Info.plist"

test:
	go test -race ./internal/ping
	npm --prefix frontend run build
