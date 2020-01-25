package main

import (
    "fmt"
    "time"
    "math/rand"
    "os"
    "strconv"
)

func runClient() {
    if len(os.Args) < 3 || os.Args[2] != "dump" {
        rand.Seed(time.Now().Unix())
        val := rand.Int()
        collectingTube.Put([]byte(strconv.Itoa(val)), 1, 0, 30 * time.Second)
        fmt.Println("Sent value:", val)
    } else {
        collectingTube.Put([]byte("dump"), 1, 0, 30 * time.Second)
        fmt.Println("Dump requested")
        id, body, err := auxiliaryTubeSet.Reserve(5 * time.Second)
        if err == nil {
            queueConn.Delete(id)
            fmt.Print(string(body))
        }
    }
}
