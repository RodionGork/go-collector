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

var maxRandom = -1

func runClient() {
    if len(os.Args) > 2 {
        cmd := os.Args[2]
        if cmd == "dump" {
            requestDump()
        } else if cmd == "stats" {
            showServerStats()
        } else if cmd == "send" {
            sendValue(parseArgOrRandom())
        } else {
            log.Errorf("Unrecognized operation: %s", cmd)
        }
    } else {
        printHelp()
    }
}

func printHelp() {
    fmt.Println("Usage:")
    fmt.Println("    ./go-collector client <stats|dump>")
    fmt.Println("    ./go-collector client send [value]")
}

func sendValue(val int) {
    data := &Command {
        Cmd: Command_PUT,
        Val: []int64 {int64(val)},
    }
    dataBin, _ := proto.Marshal(data)
    collectingTube.Put(dataBin, 1, 0, 30 * time.Second)
    log.Debugf("Sent value: %d", val)
}

func parseArgOrRandom() int {
    val := generateRandomToSend()
    if len(os.Args) > 3 {
        v, e := strconv.Atoi(os.Args[3])
        if e == nil {
            val = v
        } else {
            log.Warnf("can't parse value to send, sending random instead")
        }
    }
    return val
}

func generateRandomToSend() int {
    if maxRandom < 0 {
        rand.Seed(time.Now().UnixNano())
        maxRandom = confGetInt("maxClientRandom")
    }
    val := rand.Int()
    if maxRandom > 0 {
        val %= maxRandom
    }
    return val
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

func showServerStats() {
    rec, sto, err := requestServerStats()
    if err == nil {
        fmt.Printf("Server stats: received=%d, stored=%d\n", rec, sto)
    } else {
        log.Errorf("can't fetch stats: %s", err)
    }
}

func requestServerStats() (int64, int64, error) {
    body, err := auxiliaryRequest(Command_STATS, "Stats requested")
    if err != nil {
        return 0, 0, err
    }
    data := &Stats {}
    proto.Unmarshal(body, data)
    return data.Received, data.Stored, nil
}
