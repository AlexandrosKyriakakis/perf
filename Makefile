all: inbound outbound

clean:
	rm inbound/inbound
	rm outbound/outbound

inbound: 
	cd $@; go build .

outbound: 
	cd $@; go build .

in-run:
	go run inbound/main.go -fixconfig=cfg/inbound.cfg -samplesize=10000

out-run:
	go run outbound/main.go -fixconfig=cfg/outbound.cfg -samplesize=10000

.PHONY: inbound outbound
