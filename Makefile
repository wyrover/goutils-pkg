all:
	dep ensure -v && go build -x cmd/...

clean:
	go clean -x

test:
	go test -x cmd/...

install: all
	go install -x cmd/...

uninstall: clean
	go clean -x -i cmd/...

.PHONY: all clean test install
