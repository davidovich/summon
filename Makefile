SHELL=/bin/bash

HAS_PACKR2	:= $(shell command -v packr2)

SCAFFOLD_BIN := bin/scaffold
PACKR_FILE := pkg/scaffold/scaffold-packr.go

ASSETS := $(shell find templates/scaffold)

all: test $(PACK_FILE) $(SCAFFOLD_BIN)

.PHONY: $(SCAFFOLD_BIN)
$(SCAFFOLD_BIN): $(PACKR_FILE)
	@mkdir -p bin
	go build -o $@ $(@F)/$(@F).go

$(PACKR_FILE): $(ASSETS)
ifndef HAS_PACKR2
	go get -u github.com/gobuffalo/packr/v2/packr2
endif
	cd $(@D) && GO111MODULE=on packr2

.PHONY: test
test:
	go test -cover -v ./...

.PHONY: clean
clean:
	cd pkg/scaffold && packr2 clean
	rm -r bin
