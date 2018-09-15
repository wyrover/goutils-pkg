PROJECT =	"github.com/kisom/goutils/cmd"

all:
	dep ensure -v && go build -x $(PROJECT)/...

clean:
	go clean -x

test: all
	go test -x $(PROJECT)/...

install: all
	go install -x $(PROJECT)/...

uninstall: clean
	go clean -x -i $(PROJECT)/...

check: all
	go vet $(PROJECT)/...
	golint $(PROJECT)/...
	staticcheck $(PROJECT)/...


.PHONY: all clean test install check
