package main

import (
	"bytes"
	// Standard libraries
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"time"

	// Custom libraries
	"github.com/felixge/fgprof"
	"github.com/google/uuid"
	"github.com/grd/stat"
	"github.com/quickfixgo/enum"
	"github.com/quickfixgo/field"
	fix42er "github.com/quickfixgo/fix42/executionreport"
	"github.com/quickfixgo/quickfix"
	"github.com/shopspring/decimal"
)

var (
	allDone    = make(chan struct{})
	fixconfig  = flag.String("fixconfig", "outbound.cfg", "FIX config file")
	sampleSize = flag.Int("samplesize", 2, "Expected sample size")
)

var (
	SessionID quickfix.SessionID
	start     = make(chan interface{})
	app       = &OutboundRig{}
	metrics   stat.IntSlice
	t0        time.Time
	count     = 0
)

func main() {
	flag.Parse()

	http.DefaultServeMux.Handle("/debug/fgprof", fgprof.Handler())
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	metrics = make(stat.IntSlice, *sampleSize)

	cfg, err := os.Open(*fixconfig)
	if err != nil {
		log.Fatal(err)
	}

	appSettings, err := quickfix.ParseSettings(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// logFactory, err := quickfix.NewFileLogFactory(appSettings)
	logFactory := quickfix.NewNullLogFactory()
	if err != nil {
		log.Fatal(err)
	}

	storeFactory := quickfix.NewMemoryStoreFactory()

	initiator, err := quickfix.NewInitiator(app, storeFactory, appSettings, logFactory)
	if err != nil {
		log.Fatal(err)
	}
	if err = initiator.Start(); err != nil {
		log.Fatal(err)
	}

	<-start

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
	// send the same message multiple times
	for i := 0; i < *sampleSize; i++ {
		err := quickfix.SendToTarget(execReport, SessionID)
		if err != nil {
			log.Println(err.Error())
		}
	}

	<-allDone
	elapsed := time.Since(t0)
	metricsUS := make(stat.Float64Slice, *sampleSize)
	for i, durationNS := range metrics {
		metricsUS[i] = float64(durationNS) / 1000.0
	}

	mean := stat.Mean(metricsUS)
	max, maxIndex := stat.Max(metricsUS)
	stdev := stat.Sd(metricsUS)

	log.Printf(">>>>>>>>>>> OUTBOUND STATS <<<<<<<<<<<")
	log.Print("NumCPU: ", runtime.NumCPU())
	log.Print("GOMAXPROCS: ", runtime.GOMAXPROCS(-1))
	log.Printf("Sample mean is %g us", mean)
	log.Printf("Sample max is %g us (%v)", max, maxIndex)
	log.Printf("Standard Dev is %g us", stdev)
	log.Printf("Processed %d msg in %v [effective rate: %.4f msg/s]", count, elapsed, float64(count)/float64(elapsed)*float64(time.Second))
	log.Printf("----------- OUTBOUND STATS -----------")
}

type OutboundRig struct {
	quickfix.SessionID
}

func (e OutboundRig) OnCreate(sessionID quickfix.SessionID) {}
func (e *OutboundRig) OnLogon(sessionID quickfix.SessionID) {
	SessionID = sessionID
	start <- "START"
	t0 = time.Now()
}
func (e OutboundRig) OnLogout(sessionID quickfix.SessionID)                              {}
func (e OutboundRig) ToAdmin(msgBuilder *quickfix.Message, sessionID quickfix.SessionID) {}
func (e OutboundRig) ToApp(msgBuilder *quickfix.Message, sessionID quickfix.SessionID) (err error) {
	return
}

func (e OutboundRig) FromAdmin(msg *quickfix.Message, sessionID quickfix.SessionID) (err quickfix.MessageRejectError) {
	return
}

func (e OutboundRig) FromApp(msg *quickfix.Message, sessionID quickfix.SessionID) (err quickfix.MessageRejectError) {
	metrics[count] = int64(time.Since(msg.ReceiveTime))
	count++
	if bytes.Contains(msg.Bytes(), []byte("\x0134="+strconv.Itoa(*sampleSize)+"\x01")) {
		allDone <- struct{}{}
	}
	return
}
