package main

import (
    "fmt"
    "time"
    "math/rand"
    "os"
    "strconv"
    "github.com/beanstalkd/go-beanstalk"
)

func runClient() {
    conn, _ := beanstalk.Dial("tcp", confGet("queueHost"))
    rand.Seed(time.Now().Unix())
    val := rand.Int()
    if len(os.Args) < 3 || os.Args[2] != "dump" {
        conn.Put([]byte(strconv.Itoa(val)), 1, 0, 30 * time.Second)
        fmt.Println("Sent value:", val)
    } else {
        conn.Put([]byte("dump"), 1, 0, 30 * time.Second)
        fmt.Println("Dump requested")
    }
}
