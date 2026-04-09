.PHONY: all lib test test-race lint fmt vet check coverage clean sync sync-local

all: lib test

lib: csuzume/build/lib/libsuzume.a

csuzume/build/lib/libsuzume.a: csuzume/CMakeLists.txt
	cd csuzume && cmake -B build -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=OFF && cmake --build build -j$$(nproc 2>/dev/null || sysctl -n hw.ncpu)

csuzume/CMakeLists.txt:
	./sync-upstream.sh

test: lib
	go test ./... -count=1

test-race: lib
	go test -race ./... -count=1

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w -local github.com/libraz/go-suzume .

vet:
	go vet ./...

check: vet lint

coverage: lib
	go test -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html

clean:
	rm -rf csuzume/build coverage.txt coverage.html

sync:
	./sync-upstream.sh

sync-local:
	./sync-upstream.sh --local
