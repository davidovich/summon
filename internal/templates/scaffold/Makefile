SHELL=/bin/bash

HAS_PACKR2	:= $(shell command -v packr2;)

SUMMONER_NAME := {{ .SummonerName }}

SUMMON_DEPS = $(shell GO111MODULE=on go list -f '{{`{{ $$dir := .Dir}}{{range .GoFiles}}{{ printf "%s/%s\n" $$dir . }}{{end}}`}}' ./... github.com/davidovich/summon/...)
ASSETS := $(shell find assets)

all: $(SUMMONER_NAME)/packrd/packed-packr.go bin/$(SUMMONER_NAME)

bin/$(SUMMONER_NAME): $(SUMMON_DEPS)
	@mkdir -p bin
	go build -o $@ $(SUMMONER_NAME)/summon.go

$(SUMMONER_NAME)/packrd/packed-packr.go: $(ASSETS)
ifndef HAS_PACKR2
	go get -u github.com/gobuffalo/packr/v2/packr2
endif
	cd $(SUMMONER_NAME) && GO111MODULE=on packr2

.PHONY: clean
clean:
	cd $(SUMMONER_NAME) && packr2 clean
	rm -rf bin