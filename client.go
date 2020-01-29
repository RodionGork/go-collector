package main

import (
    "fmt"
    "time"
    "math/rand"
    "os"
    "strconv"
    "github.com/golang/protobuf/proto"
    log "github.com/sirupsen/logrus"
)

var maxRandom int

func runClient() {
    maxRandom = confGetInt("maxClientRandom")
    if len(os.Args) > 2 {
        cmd := os.Args[2]
        if cmd == "dump" {
            requestDump()
        } else if cmd == "stats" {
            requestStats()
        } else if cmd == "send" {
            sendValue()
        } else {
            log.Errorf("Unrecognized operation: %s", cmd)
        }
    } else {
        printHelp()
    }
}

func init() {
    rand.Seed(time.Now().UnixNano())
}

func printHelp() {
    fmt.Println("Usage:")
    fmt.Println("    ./go-collector client <stats|dump>")
    fmt.Println("    ./go-collector client send [value]")
}

func sendValue() {
    val := rand.Int()
    if maxRandom > 0 {
        val %= maxRandom
    }
    if len(os.Args) > 3 {
        v, e := strconv.Atoi(os.Args[3])
        if e == nil {
            val = v
        } else {
            log.Warnf("can't parse value to send, sending random instead")
        }
    }
    data := &Command {
        Cmd: Command_PUT,
        Val: []int64 {int64(val)},
    }
    dataBin, _ := proto.Marshal(data)
    collectingTube.Put(dataBin, 1, 0, 30 * time.Second)
    log.Infof("Sent value: %d", val)
}

func cleanAuxiliaryTube() {
    for true {
        id, _, err := responseTubeSet.Reserve(0 * time.Second)
        if err == nil {
            queueConn.Delete(id)
        } else {
            break
        }
    }
}

func auxiliaryRequest(cmd Command_Code, msg string) ([]byte, error) {
    cleanAuxiliaryTube()
    data := &Command { Cmd: cmd }
    dataBin, _ := proto.Marshal(data)
    auxiliaryTube.Put(dataBin, 1, 0, 30 * time.Second)
    log.Info(msg)
    id, body, err := responseTubeSet.Reserve(5 * time.Second)
    if err == nil {
        queueConn.Delete(id)
    } else {
        log.Warnf("response retrieval failed!")
    }
    return body, err
}

func requestDump() {
    body, err := auxiliaryRequest(Command_DUMP, "Dump requested")
    if err == nil {
        data := &Dump {}
        proto.Unmarshal(body, data)
        for _, v := range data.List {
            fmt.Println(v.Val, v.Cnt)
        }
    }
}

func requestStats() {
    body, err := auxiliaryRequest(Command_STATS, "Stats requested")
    if err == nil {
        data := &Stats {}
        proto.Unmarshal(body, data)
        fmt.Printf("Server stats: received=%d, stored=%d\n", data.Received, data.Stored)
    }
}
