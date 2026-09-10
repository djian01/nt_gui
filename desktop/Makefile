.PHONY: setup frontend icons windows-icons build run test mac-app linux-package

TARGET_OS := $(shell go env GOOS)
TARGET_ARCH := $(shell go env GOARCH)
EXECUTABLE := bin/net-test
ifeq ($(TARGET_OS),windows)
EXECUTABLE := bin/net-test.exe
WINDOWS_LDFLAGS := -ldflags "-H windowsgui"
build: windows-icons
endif

.DEFAULT_GOAL := setup

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

icons:
	npm --prefix frontend run icons

windows-icons:
	go run github.com/akavel/rsrc@v0.10.2 -arch $(TARGET_ARCH) -ico build/icons/net-test.ico -o rsrc_windows_$(TARGET_ARCH).syso

build: frontend
	go build -tags production $(WINDOWS_LDFLAGS) -o $(EXECUTABLE) .

run: build
	./$(EXECUTABLE)

mac-app: build
	mkdir -p "bin/NET-Test.app/Contents/MacOS" "bin/NET-Test.app/Contents/Resources"
	cp bin/net-test "bin/NET-Test.app/Contents/MacOS/net-test"
	cp build/darwin/Info.plist "bin/NET-Test.app/Contents/Info.plist"
	cp build/icons/net-test.icns "bin/NET-Test.app/Contents/Resources/net-test.icns"
	cp ../LICENSE ../THIRD_PARTY_NOTICES.txt "bin/NET-Test.app/Contents/Resources/"

linux-package: build
	mkdir -p bin/net-test-linux/usr/bin bin/net-test-linux/usr/share/applications
	mkdir -p bin/net-test-linux/usr/share/icons/hicolor/scalable/apps bin/net-test-linux/usr/share/icons/hicolor/512x512/apps
	cp bin/net-test bin/net-test-linux/usr/bin/net-test
	cp build/linux/net-test.desktop bin/net-test-linux/usr/share/applications/net-test.desktop
	cp frontend/public/net-test.svg bin/net-test-linux/usr/share/icons/hicolor/scalable/apps/net-test.svg
	cp build/icons/net-test.png bin/net-test-linux/usr/share/icons/hicolor/512x512/apps/net-test.png
	mkdir -p bin/net-test-linux/usr/share/licenses/net-test
	cp ../LICENSE ../THIRD_PARTY_NOTICES.txt bin/net-test-linux/usr/share/licenses/net-test/

test: frontend
	go test -race ./...
