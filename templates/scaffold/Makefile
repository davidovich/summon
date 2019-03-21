SHELL=/bin/bash

HAS_PACKR2	:= $(shell command -v packr2)

SUMMONER_NAME := {{ .SummonerName }}

SUMMON_DEPS = $(shell find $(SUMMONER_NAME))
ASSETS := $(shell find assets)

all: $(SUMMONER_NAME)/packrd/packed-packr.go bin/$(SUMMONER_NAME)

bin/$(SUMMONER_NAME): $(SUMMON_DEPS)
	@mkdir -p bin
	cd $(SUMMONER_NAME) && go build -o ../$@ summon.go

$(SUMMONER_NAME)/packrd/packed-packr.go: $(ASSETS)
ifndef HAS_PACKR2
	go get -u github.com/gobuffalo/packr/v2/packr2
endif
	cd $(SUMMONER_NAME) && GO111MODULE=on packr2

.PHONY: clean
clean:
	cd $(SUMMONER_NAME) && packr2 clean
