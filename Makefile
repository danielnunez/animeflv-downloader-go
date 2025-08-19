GOOS ?= darwin
GOARCH ?= amd64
GOARM ?= 7

.PHONY: all linux macos-amd64 windows clean

all: linux macos-amd64 windows

linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux/animeflv-downloader .

macos-amd64:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/macos/animeflv-downloader .

windows:
	GOOS=windows GOARCH=amd64 go build -o ./bin/windows/animeflv-downloader.exe .

clean:
	rm -rf ./bin/
