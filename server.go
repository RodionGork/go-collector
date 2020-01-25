package main

import (
    "bytes"
    "fmt"
    "time"
    "os"
    "os/signal"
    "strconv"
    "syscall"
)

var divisor, remainder int

var valuesChan chan int

var storage = map[int]int {}


func runServer() {
    
    initAndSetup()
    
    serveEndlessly()
}

func initAndSetup() {
    divisor = confGetInt("divisor")
    remainder = confGetInt("remainder")

    fmt.Println("Starting server, press Ctrl-C to exit...")
    setCtrlC()
    
    valuesChan = make(chan int)
    go collector(valuesChan);
}

func collector(ch chan int) {
    for val := range ch {
        fmt.Println("Storing", val)
        storage[val]++
    }
}

func serveEndlessly() {
    for true {
        id, body, err := collectingTubeSet.Reserve(5 * time.Second)
        if (err == nil) {
            sBody := string(body)
            val, err := strconv.Atoi(sBody)
            if err == nil {
                fmt.Println("Received", id, val)
                if checkValue(val) {
                    storeValue(val)
                }
            } else if (sBody == "dump") {
                dumpValues()
            }
            queueConn.Delete(id)
        }
    }
}

func checkValue(val int) bool {
    return val % divisor == remainder
}

func storeValue(val int) {
    valuesChan <- val
}

func dumpValues() {
    var buf bytes.Buffer
    for k, v := range storage {
        buf.WriteString(fmt.Sprintf("%d:%d\n", k, v))
    }
    auxiliaryTube.Put(buf.Bytes(), 1, 0, 30 * time.Second)
    fmt.Println("Dump sent:", len(storage), "values")
}

func setCtrlC() {
    ch := make(chan os.Signal, 2)
    signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
    go func() {
        <- ch
        fmt.Println("Ctrl-C caught, exiting")
        os.Exit(0)
    }()
}
