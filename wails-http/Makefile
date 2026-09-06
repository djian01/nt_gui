.PHONY: setup frontend build run test mac-app

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
