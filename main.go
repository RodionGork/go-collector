package main

import (
    "flag"
    "fmt"
    "os"
    "io/ioutil"
    "encoding/json"
    "github.com/beanstalkd/go-beanstalk"
)

var queueConn *beanstalk.Conn
var collectingTube *beanstalk.Tube
var auxiliaryTube *beanstalk.Tube
var collectingTubeSet *beanstalk.TubeSet
var auxiliaryTubeSet *beanstalk.TubeSet

var parsedConfig map[string]interface{}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("USAGE: go <server|client>")
    } else if os.Args[1] == "server" {
        runServer()
    } else if os.Args[1] == "client" {
        runClient()
    } else {
        fmt.Println("Unsupported execution mode:", os.Args[1])
        os.Exit(1)
    }
}

func init() {
    initConfig()
    if flag.Lookup("test.v") == nil {
        initConnection()
    } else {
        fmt.Println("Skipping queue connection")
    }
}

func initConfig() {
    data, err := ioutil.ReadFile("config.json")
    if err != nil {
        fmt.Println("Error: config.json not found!")
        os.Exit(10)
    }
    var jData interface{}
    err2 := json.Unmarshal(data, &jData)
    if err2 != nil {
        fmt.Println("Error: config.json parse failure!")
        os.Exit(11)
    }
    parsedConfig = jData.(map[string]interface{})
}

func initConnection() {
    conn, err := beanstalk.Dial("tcp", confGet("queueHost"))
    if err != nil {
        fmt.Println("Can't connect to message queue")
        os.Exit(12)
    }
    queueConn = conn
    collectingTube = &beanstalk.Tube { Conn: conn, Name: "collector-tube-in" }
    collectingTubeSet = beanstalk.NewTubeSet(conn, collectingTube.Name)
    auxiliaryTube = &beanstalk.Tube { Conn: conn, Name: "collector-tube-out" }
    auxiliaryTubeSet = beanstalk.NewTubeSet(conn, auxiliaryTube.Name)
}

func confGet(key string) string {
    val, ok := parsedConfig[key]
    if ok {
        s, ok := val.(string)
        if ok {
            return s
        }
    }
    return ""
}

func confGetInt(key string) int {
    val, ok := parsedConfig[key]
    if ok {
        f, ok := val.(float64)
        if ok {
            return int(f)
        }
    }
    return 0
}
