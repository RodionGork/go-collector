package main

import (
    "fmt"
    "time"
    "math/rand"
    "strconv"
    "github.com/beanstalkd/go-beanstalk"
)

func main() {
    conn, _ := beanstalk.Dial("tcp", "127.0.0.1:11300")
    rand.Seed(time.Now().Unix())
    val := rand.Int()
    conn.Put([]byte(strconv.Itoa(val)), 1, 0, 30 * time.Second)
    fmt.Println("Sent value:", val)
}
