PREFIX ?=	$(HOME)

all:
	gb build

clean:
	rm -rf bin pkg

test:
	gb test

install: all
	cp bin/* $(PREFIX)/bin/

.PHONY: all clean test install
