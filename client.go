package main

import (
    "fmt"
    "time"
    "math/rand"
    "os"
    "strconv"
    "github.com/golang/protobuf/proto"
)

func runClient() {
    if len(os.Args) < 3 || os.Args[2] != "dump" {
        sendValue()
    } else {
        requestDump()
    }
}

func sendValue() {
    rand.Seed(time.Now().Unix())
    val := rand.Int()
    if len(os.Args) >= 3 {
        v, e := strconv.Atoi(os.Args[2])
        if e == nil {
            val = v
        }
    }
    data := &Command {
        Cmd: Command_PUT,
        Val: []int64 {int64(val)},
    }
    dataBin, _ := proto.Marshal(data)
    collectingTube.Put(dataBin, 1, 0, 30 * time.Second)
    fmt.Println("Sent value:", val)
}

func requestDump() {
    data := &Command { Cmd: Command_DUMP }
    dataBin, _ := proto.Marshal(data)
    collectingTube.Put(dataBin, 1, 0, 30 * time.Second)
    fmt.Println("Dump requested")
    id, body, err := auxiliaryTubeSet.Reserve(5 * time.Second)
    if err == nil {
        queueConn.Delete(id)
        data := &Dump {}
        proto.Unmarshal(body, data)
        for _, v := range data.List {
            fmt.Println(v.Val, v.Cnt)
        }
    }
}
