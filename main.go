package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "encoding/json"
)

var parsed map[string]interface{}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("USAGE: go <server|client>")
    } else if os.Args[1] == "server" {
        RunServer()
    } else if os.Args[1] == "client" {
        RunClient()
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
    parsed = jData.(map[string]interface{})
}

func ConfGet(key string) string {
    return parsed[key].(string)
}

func ConfGetInt(key string) int {
    return int(parsed[key].(float64))
}
