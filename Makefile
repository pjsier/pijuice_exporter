VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
LDFLAGS := -ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

.PHONY: clean start format lint build

start:
	go run ./main.go file

format:
	test -z $$(gofmt -l .)

lint:
	golangci-lint run

clean:
	rm -rf release

build: release/pijuice_exporter-linux-arm.tar.gz

release/pijuice_exporter-linux-arm.tar.gz: release/pijuice_exporter-linux-arm
	tar -czvf $@ $<

release/pijuice_exporter-linux-arm:
	GOOS=linux GOARCH=arm GOARM=5 go build $(LDFLAGS) -o $@/pijuice_exporter ./main.go
