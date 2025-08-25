
UNAME_S := $(shell uname)

ifeq ($(origin FLOXBIN), undefined)
  ifeq ($(UNAME_S), Darwin)
    FLOXBIN := /usr/local/bin/flox
  else ifeq ($(UNAME_S), Linux)
    FLOXBIN := /usr/bin/flox
  else
    $(error Unsupported OS: $(UNAME_S))
  endif
endif

all: build

fmt:
	go fmt ./...

tidy: fmt
	go mod tidy

build: fmt
	go build .

clean:
	rm -f quotes-app-go
	@$(FLOXBIN) build clean
