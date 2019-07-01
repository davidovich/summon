SHELL=/bin/bash

HAS_PACKR2				:= $(shell command -v packr2)
HAS_GOBIN				:= $(shell command -v gobin)
HAS_GOCOVERUTIL			:= $(shell command -v gocoverutil)

ifndef HAS_GOBIN
$(shell GO111MODULE=off go get -u github.com/myitcv/gobin 2>/dev/null)
endif

export GO111MODULE := on

SCAFFOLD_BIN := bin/scaffold
PACKR_FILE := pkg/scaffold/scaffold-packr.go
COVERAGE_PERCENT_FILE := build/coverage-percent.txt

ASSETS := $(shell find templates/scaffold)

all: $(PACKR_FILE) test $(SCAFFOLD_BIN)

.PHONY: bin
bin: $(PACKR_FILE) $(SCAFFOLD_BIN)

.PHONY: $(SCAFFOLD_BIN)
$(SCAFFOLD_BIN): $(PACKR_FILE)
	@mkdir -p bin
	go build -o $@ $(@F)/$(@F).go

$(PACKR_FILE): $(ASSETS)
ifndef HAS_PACKR2
	gobin github.com/gobuffalo/packr/v2/packr2
endif
	packr2

GO_TESTS := pkg internal cmd
COVERAGE := $(foreach t,$(GO_TESTS),build/coverage/report/$(t))
MERGED_COVERAGE := build/coverage/report/cover.merged.out
HTML_COVERAGE := build/coverage/html/index.html

.PHONY: test
test: output-coverage

.PHONY: output-coverage
output-coverage: $(MERGED_COVERAGE) $(HTML_COVERAGE)
	go tool cover -func=$< | sed -e 's/total:[[:space:]]*(statements)[[:space:]]*\([0-9.]*\)%/\1/gw $(COVERAGE_PERCENT_FILE)'

$(MERGED_COVERAGE): $(COVERAGE)
ifndef HAS_GOCOVERUTIL
	gobin github.com/AlekSi/gocoverutil@v0.2.0
endif
	$(call msg,Generating merged coverage report)
	gocoverutil -coverprofile=$@ merge $^

$(HTML_COVERAGE): $(MERGED_COVERAGE)
	@mkdir -p $(@D)
	go tool cover -html=$< -o $@

.PHONY: $(COVERAGE)
$(COVERAGE):
	@mkdir -p $(@D)
	$(call msg,--> Testing $(@F)...)
	go test ./$(@F)/... --cover -coverprofile $@ -v

.PHONY: clean
clean:
	packr2 clean
	rm -r bin build
