GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)

all: silo

silo: run

dependencies:
	go get github.com/unixvoid/glogger
	go get gopkg.in/gcfg.v1
	go get gopkg.in/redis.v5

run:
	go run \
		silo/silo.go \
		silo/populate_packages.go

stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/silo silo/*.go

populate_test:
	mkdir -p rkt/nsproxy/
	wget -O rkt/nsproxy/nsproxy-latest-linux-amd64.aci https://cryo.unixvoid.com/bin/rkt/nsproxy/nsproxy-latest-linux-amd64.aci
	wget -O rkt/nsproxy/nsproxy-latest-linux-amd64.aci.asc https://cryo.unixvoid.com/bin/rkt/nsproxy/nsproxy-latest-linux-amd64.aci.asc
	mkdir -p rkt/cryodns/
	wget -O rkt/cryodns/cryodns-latest-linux-amd64.aci https://cryo.unixvoid.com/bin/rkt/cryodns/cryodns-latest-linux-amd64.aci
	wget -O rkt/cryodns/cryodns-latest-linux-amd64.aci.asc https://cryo.unixvoid.com/bin/rkt/cryodns/cryodns-latest-linux-amd64.aci.asc
	mkdir -p rkt/binder/
	wget -O rkt/binder/binder-latest-linux-amd64.aci https://cryo.unixvoid.com/bin/rkt/binder/standalone/binder-latest-linux-amd64.aci
	wget -O rkt/binder/binder-latest-linux-amd64.aci.asc https://cryo.unixvoid.com/bin/rkt/binder/standalone/binder-latest-linux-amd64.aci.asc
	mkdir -p rkt/pubkey/
	wget -O rkt/pubkey/pubkey.gpg https://cryo.unixvoid.com/bin/rkt/pubkey/pubkeys.gpg

clean:
	rm -rf bin/

cleanall: clean
	rm -rf rkt/
