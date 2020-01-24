package main

import (
    "fmt"
    "time"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "github.com/beanstalkd/go-beanstalk"
)

var queueConn *beanstalk.Conn

var divisor, remainder int

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

    conn, err := beanstalk.Dial("tcp", confGet("queueHost"))
    if err != nil {
        fmt.Println("Can't connect to message queue")
        os.Exit(1)
    }
    queueConn = conn
}

func serveEndlessly() {
    for true {
        id, body, err := queueConn.Reserve(5 * time.Second)
        if (err == nil) {
            sBody := string(body)
            val, err := strconv.Atoi(sBody)
            if err == nil {
                fmt.Println("Received", id, val)
                if checkValue(val) {
                    storeValue(id, val)
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

func storeValue(id uint64, val int) {
    fmt.Println("Storing", id, val)
    storage[val]++
}

func dumpValues() {
    fmt.Println("Dump:", len(storage), "values")
    for k, v := range storage {
        fmt.Println("D", k, v)
    }
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
