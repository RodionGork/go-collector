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

var divisor, remainder int

var valuesChan chan int

var storage = map[int]int {}


func runServer() {
    
    initAndSetup()
    
    go serveEndlessly(auxiliaryTubeSet, processAuxiliaryCommands)
    serveEndlessly(collectingTubeSet, processIncomingValues)
}

func initAndSetup() {
    divisor = confGetInt("divisor")
    remainder = confGetInt("remainder")

    log.Info("Starting server, use Ctrl-C to exit")
    setCtrlC()
    
    valuesChan = make(chan int)
    go collector(valuesChan);
}

func collector(ch chan int) {
    for val := range ch {
        log.Info("Storing", val) //todo: pass ID here
        storage[val]++
    }
}

func serveEndlessly(tubeSet *beanstalk.TubeSet, processor func(id uint64, body []byte)) {
    for true {
        id, body, err := tubeSet.Reserve(0 * time.Second)
        if (err == nil) {
            processor(id, body)
        }
    }
}

func processIncomingValues(id uint64, body []byte) {

    defer func() { queueConn.Delete(id) } ()
    defer processingErrorCheck(id);

    var cmd = &Command {}
    proto.Unmarshal(body, cmd)
    if cmd.Cmd == Command_PUT {
        val := int(cmd.Val[0])
        log.Infof("Received, msgid=%d, value=%d", id, cmd.Val)
        if checkValue(val) {
            storeValue(val)
        }
    } else {
        panic("Unexpected command in incoming tube")
    }
}

func processAuxiliaryCommands(id uint64, body []byte) {

    defer func() { queueConn.Delete(id) } ()
    defer processingErrorCheck(id);

    var cmd = &Command {}
    proto.Unmarshal(body, cmd)
    if (cmd.Cmd == Command_DUMP) {
        dumpValues()
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

func storeValue(val int) {
    valuesChan <- val
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

func setCtrlC() {
    ch := make(chan os.Signal, 2)
    signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- ch
        log.Info("Ctrl-C caught, exiting")
        os.Exit(0)
    }()
}
