package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "encoding/json"
)

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
    }
}

func init() {
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
