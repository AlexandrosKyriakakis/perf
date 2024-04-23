module github.com/quickfixgo/perf

go 1.22.2

require (
	github.com/grd/stat v0.0.0-20130623202159-138af3fd5012
	github.com/quickfixgo/enum v0.1.0
	github.com/quickfixgo/field v0.1.0
	github.com/quickfixgo/fix42 v0.1.0
	github.com/quickfixgo/quickfix v0.9.0
)

require github.com/google/pprof v0.0.0-20211214055906-6f57359322fd // indirect

require (
	github.com/armon/go-proxyproto v0.1.0 // indirect
	github.com/felixge/fgprof v0.9.3
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/quickfixgo/tag v0.1.0 // indirect
	github.com/shopspring/decimal v1.3.1
	golang.org/x/net v0.20.0 // indirect
)

replace github.com/quickfixgo/quickfix v0.9.0 => github.com/alpacahq/quickfix v0.6.1-0.20240201125837-6eaa8ce00453
