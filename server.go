package main

import (
    "time"
    "os"
    "os/signal"
    "syscall"
    "github.com/beanstalkd/go-beanstalk"
    "github.com/golang/protobuf/proto"
    log "github.com/sirupsen/logrus"
)

type idAndVal struct {
    id uint64
    val int
}

var divisor, remainder int
var receiveSpeedLimit int

var valuesChan chan *idAndVal
var metricsChan chan bool

var storage = map[int]int {}
var metricReceived, metricStored int64

func runServer() {
    
    initAndSetup()
    
    go serveEndlessly(auxiliaryTubeSet, 5, processAuxiliaryCommands)
    serveEndlessly(collectingTubeSet, receiveSpeedLimit, processIncomingValues)
}

func initAndSetup() {
    divisor = confGetInt("divisor")
    remainder = confGetInt("remainder")
    receiveSpeedLimit = confGetInt("receivePerSecondMax")

    log.Info("Starting server, use Ctrl-C to exit")
    setCtrlC()
    
    valuesChan = make(chan *idAndVal)
    go collector()
    metricsChan = make(chan bool)
    go metricProcessor()
}

func collector() {
    for next := range valuesChan {
        log.Infof("Storing %d (id=%d)", next.val, next.id) //todo: pass ID here
        storage[next.val]++
    }
}

func serveEndlessly(tubeSet *beanstalk.TubeSet, maxSpeed int, processor func(id uint64, body []byte)) {
    rateLimiter(&maxSpeed, func() {
        id, body, err := tubeSet.Reserve(0 * time.Second)
        if (err == nil) {
            processor(id, body)
        }
    })
}

func processIncomingValues(id uint64, body []byte) {

    defer func() { queueConn.Delete(id) } ()
    defer processingErrorCheck(id);

    var cmd = &Command {}
    proto.Unmarshal(body, cmd)
    if cmd.Cmd == Command_PUT {
        val := int(cmd.Val[0])
        log.Debugf("Received, msgid=%d, value=%d", id, cmd.Val)
        acceptable := checkValue(val)
        if acceptable {
            data := &idAndVal { id: id, val: val }
            valuesChan <- data
        }
        metricsChan <- acceptable
    } else {
        panic("Unexpected command in incoming tube")
    }
}

func processAuxiliaryCommands(id uint64, body []byte) {

    defer func() { queueConn.Delete(id) } ()
    defer processingErrorCheck(id);

    var cmd = &Command {}
    proto.Unmarshal(body, cmd)
    if cmd.Cmd == Command_DUMP {
        dumpValues()
    } else if cmd.Cmd == Command_STATS {
        sendMetrics()
    } else {
        panic("Unexpected command in auxiliary tube")
    }
}

func processingErrorCheck(id uint64) {
    if r := recover(); r != nil {
        log.Warnf("Processing failed for id=%d with message: %s", id, r.(error).Error())
    }
}

func checkValue(val int) bool {
    return val % divisor == remainder
}

func dumpValues() {
    var dump = &Dump { List: []*Dump_ValAndCnt {} }
    for k, v := range storage {
        valAndCnt := &Dump_ValAndCnt {
            Val: int64(k),
            Cnt: int32(v),
        }
        dump.List = append(dump.List, valAndCnt)
    }
    bin, _ := proto.Marshal(dump)
    responseTube.Put(bin, 1, 0, 30 * time.Second)
    log.Infof("Dump sent: %d values", len(storage))
}

func sendMetrics() {
    var stats = &Stats {
        Received: metricReceived,
        Stored: metricStored,
    }
    bin, _ := proto.Marshal(stats)
    responseTube.Put(bin, 1, 0, 30 * time.Second)
    log.Infof("Metrics sent: %d received, %d stored", metricReceived, metricStored)
}

func metricProcessor() {
    for stored := range metricsChan {
        metricReceived++
        if stored {
            metricStored++
        }
    }
}

func setCtrlC() {
    ch := make(chan os.Signal, 2)
    signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- ch
        log.Info("Ctrl-C caught, exiting")
        os.Exit(0)
    }()
}
