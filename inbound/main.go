package main

import (
	// Standard libraries
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	// Custom libraries
	"github.com/google/uuid"
	"github.com/grd/stat"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/quickfix"
	"github.com/shopspring/decimal"

	fix42er "github.com/quickfixgo/fix42/executionreport"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	fixconfig  = flag.String("fixconfig", "inbound.cfg", "FIX config file")
	sampleSize = flag.Int("samplesize", 2, "Expected sample size")
)

var (
	count   = 0
	allDone = make(chan interface{})
	app     = &InboundRig{
		MessageRouter: quickfix.NewMessageRouter(),
	}
	metrics stat.IntSlice
	t0      time.Time
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	log.Print("NumCPU: ", runtime.NumCPU())
	log.Print("GOMAXPROCS: ", runtime.GOMAXPROCS(-1))

	metrics = make(stat.IntSlice, *sampleSize)

	cfg, err := os.Open(*fixconfig)
	if err != nil {
		log.Fatal(err)
	}

	appSettings, err := quickfix.ParseSettings(cfg)
	if err != nil {
		log.Fatal(err)
	}

	logFactory := quickfix.NewNullLogFactory()
	storeFactory := quickfix.NewMemoryStoreFactory()
	app.AddRoute(fix42er.Route(app.OnFIX42ExecutionReport))

	acceptor, err := quickfix.NewAcceptor(app, storeFactory, appSettings, logFactory)
	if err != nil {
		log.Fatal(err)
	}
	if err = acceptor.Start(); err != nil {
		log.Fatal(err)
	}

	<-allDone
	time.Sleep(500 * time.Millisecond) // Inbound metrics might be off
	elapsed := time.Since(t0)
	metricsUS := make(stat.Float64Slice, *sampleSize)
	for i, durationNS := range metrics {
		metricsUS[i] = float64(durationNS) / 1000.0
	}

	mean := stat.Mean(metricsUS)
	max, maxIndex := stat.Max(metricsUS)
	stdev := stat.Sd(metricsUS)

	log.Printf(">>>>>>>>>>> INBOUND STATS <<<<<<<<<<<")
	log.Print("NumCPU: ", runtime.NumCPU())
	log.Print("GOMAXPROCS: ", runtime.GOMAXPROCS(-1))
	log.Printf("Sample mean is %g us", mean)
	log.Printf("Sample max is %g us (%v)", max, maxIndex)
	log.Printf("Standard Dev is %g us", stdev)
	log.Printf("Processed %d msg in %v [effective rate: %.4f msg/s]", count, elapsed, float64(count)/float64(elapsed)*float64(time.Second))
	log.Printf("----------- INBOUND STATS -----------")
}

type InboundRig struct {
	*quickfix.MessageRouter
}

func (e InboundRig) OnCreate(sessionID quickfix.SessionID) {}
func (e InboundRig) OnLogon(sessionID quickfix.SessionID) {
	t0 = time.Now()
}
func (e InboundRig) OnLogout(sessionID quickfix.SessionID)                              {}
func (e InboundRig) ToAdmin(msgBuilder *quickfix.Message, sessionID quickfix.SessionID) {}
func (e InboundRig) ToApp(msgBuilder *quickfix.Message, sessionID quickfix.SessionID) (err error) {
	return
}

func (e InboundRig) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) (err quickfix.MessageRejectError) {
	return
}

func (e *InboundRig) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (err quickfix.MessageRejectError) {
	metrics[count] = int64(time.Since(msg.ReceiveTime))
	count++
	if err := e.Route(msg, sessionID); err != nil {
		return quickfix.NewMessageRejectError(err.Error(), 0, nil)
	}

	if count == *sampleSize {
		allDone <- "DONE"
	}

	return
}

func (e *InboundRig) OnFIX42ExecutionReport(msg fix42er.ExecutionReport, sessionID quickfix.SessionID) quickfix.MessageRejectError {
	execReport := fix42er.New(
		field.NewOrderID(uuid.New().String()),
		field.NewExecID(uuid.New().String()),
		field.NewExecTransType(enum.ExecTransType_NEW),
		field.NewExecType(enum.ExecType_FILL),
		field.NewOrdStatus(enum.OrdStatus_FILLED),
		field.NewSymbol("RANDOM"),
		field.NewSide("buy"),
		field.NewLeavesQty(decimal.Zero, 2),
		field.NewCumQty(decimal.RequireFromString("10"), 2),
		field.NewAvgPx(decimal.RequireFromString("10"), 2),
	)
	sendErr := quickfix.SendToTarget(execReport, sessionID)
	if sendErr != nil {
		return quickfix.NewMessageRejectError(sendErr.Error(), 0, nil)
	}

	return nil
}
