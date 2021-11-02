SUMMONER_NAME := {{ .SummonerName }}

ASSETS := $(shell find assets)

all: bin/$(SUMMONER_NAME)

bin/$(SUMMONER_NAME): summon.go $(ASSETS)
	go build -o $@ $<

.PHONY: clean
clean:
	rm -rf bin
