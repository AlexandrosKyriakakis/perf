# Performance Testing Frameworks for QuickFIX/Go

### Running

1. Start acceptor 
```bash
go run inbound/main.go --fixconfig cfg/inbound.cfg  --samplesize 100000
```
2. Start initiator
```bash
go run outbound/main.go --fixconfig cfg/outbound.cfg --samplesize 100000
```
3. Track the performance metrics 

OR

1. Build executables 
```bash
make all
```
2. Run performance test
```bash
./run-inbound-perf.sh
```