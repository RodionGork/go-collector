package main

import (
    "os"
    "time"
    "io/ioutil"
    "encoding/json"
    "github.com/beanstalkd/go-beanstalk"
    log "github.com/sirupsen/logrus"
)

var queueConn *beanstalk.Conn
var collectingTube *beanstalk.Tube
var auxiliaryTube *beanstalk.Tube
var responseTube *beanstalk.Tube
var collectingTubeSet *beanstalk.TubeSet
var auxiliaryTubeSet *beanstalk.TubeSet
var responseTubeSet *beanstalk.TubeSet

var parsedConfig map[string]interface{}

func main() {
    if len(os.Args) < 2 {
        log.Error("USAGE: ./go-collector <server|client>")
        os.Exit(1)
    } else if os.Args[1] == "server" {
        runServer()
    } else if os.Args[1] == "client" {
        runClient()
    } else if os.Args[1] == "http-client" {
        runHttpClient()
    } else {
        log.Error("Unsupported execution mode:", os.Args[1])
        os.Exit(3)
    }
}

func init() {
    initLogger()
    initConfig()
    if !isInTest() {
        initConnection()
    } else {
        log.Info("Skipping queue connection in test")
    }
}

func initLogger() {
    levelMap := map[string]log.Level {
        "debug": log.DebugLevel,
        "info": log.InfoLevel,
        "warn": log.WarnLevel,
        "error": log.ErrorLevel,
        "": log.InfoLevel,
    }
    envVal := os.Getenv("LOG_LEVEL")
    level := levelMap[envVal]
    log.SetOutput(os.Stdout)
    log.SetLevel(level)
    if envVal == "" {
        log.Infof("Default LOG_LEVEL is 'info'")
    }
}

func initConfig() {
    data, err := ioutil.ReadFile("config.json")
    if err != nil {
        log.Error("config.json not found")
        os.Exit(10)
    }
    var jData interface{}
    err2 := json.Unmarshal(data, &jData)
    if err2 != nil {
        log.Error("config.json parse failure")
        os.Exit(11)
    }
    parsedConfig = jData.(map[string]interface{})
}

func initConnection() {
    conn, err := beanstalk.Dial("tcp", confGet("queueHost"))
    if err != nil {
        log.Error("Can't connect to message queue")
        os.Exit(12)
    }
    queueConn = conn
    collectingTube = &beanstalk.Tube { Conn: conn, Name: "collector-tube" }
    collectingTubeSet = beanstalk.NewTubeSet(conn, collectingTube.Name)
    auxiliaryTube = &beanstalk.Tube { Conn: conn, Name: "auxiliary-tube" }
    auxiliaryTubeSet = beanstalk.NewTubeSet(conn, auxiliaryTube.Name)
    responseTube = &beanstalk.Tube { Conn: conn, Name: "response-tube" }
    responseTubeSet = beanstalk.NewTubeSet(conn, responseTube.Name)
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

func currentMillis() int {
    return int(time.Now().Unix() / 1000000)
}

func isInTest() bool {
    var prgName = os.Args[0]
    return prgName[len(prgName) - 5:] == ".test"
}

func rateLimiter(maxPerSec *int, f func()) {
    ts := currentMillis()
    count := 0
    for true {
        f()
        count++
        if count >= *maxPerSec {
            ts2 := currentMillis()
            delta := ts2 - ts
            if delta < 1000 {
                dur := time.Duration(1000 - delta) * time.Millisecond
                time.Sleep(dur)
                log.Tracef("Sleeping for %d ms", dur)
                ts = ts2
            }
            count = 0
        }
    }
}
