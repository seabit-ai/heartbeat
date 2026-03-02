BINARY   := heartbeat
MODULE   := github.com/seabit-ai/heartbeat
CMD      := ./go/cmd/heartbeat
DIST     := dist

LDFLAGS  := -ldflags "-s -w"

.PHONY: all build-all linux-amd64 linux-arm64 linux-arm darwin-arm64 clean

all: build-all

build-all: linux-amd64 linux-arm64 linux-arm darwin-arm64

linux-amd64:
	@mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-amd64 $(CMD)

linux-arm64:
	@mkdir -p $(DIST)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-arm64 $(CMD)

linux-arm:
	@mkdir -p $(DIST)
	GOOS=linux GOARCH=arm GOARM=7 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-arm $(CMD)

darwin-arm64:
	@mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-arm64 $(CMD)

clean:
	rm -rf $(DIST)
