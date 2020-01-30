package main

import (
    "fmt"
    "net/http"
    "io"
    "strings"
    "strconv"
    log "github.com/sirupsen/logrus"
)

var sentCount = 0
var sendRate int

func runWebClient() {
    go spamEndlessly()
    runWebInterface()
}

func spamEndlessly() {
    sendRate = confGetInt("sendPerSecondMax")
    rateLimiter(&sendRate, func() {
        sendValue(generateRandomToSend())
        sentCount++
    })
}

func runWebInterface() {
    listenAddr := ":" + strconv.Itoa(confGetInt("httpClientPort"))
    http.HandleFunc("/", webHandler)
    log.Infof("Going to serve HTTP at 127.0.0.1%s", listenAddr)
    err := http.ListenAndServe(listenAddr, nil)
    log.Errorf("server start failed with: %s", err)
}

func webHandler(resp http.ResponseWriter, req *http.Request) {
    path := req.URL.Path[1:]
    log.Infof("got request: %s", path)
    if strings.ContainsRune(path, '.') {
        http.ServeFile(resp, req, "static/" + path)
    } else if path == "" {
        http.ServeFile(resp, req, "static/index.html")
    } else if path == "client-stats" {
        io.WriteString(resp, strconv.Itoa(sentCount))
    } else if path == "server-stats" {
        serveServerStats(resp)
    } else if strings.HasPrefix(path, "client-speed/") {
        changeSpeed(path, resp)
    } else {
        io.WriteString(resp, "Here'll be dragons")
    }
}

func serveServerStats(resp http.ResponseWriter) {
    rec, sto, err := requestServerStats()
    s := "error error"
    if err == nil {
        s = fmt.Sprintf("%d %d", rec, sto)
    }
    io.WriteString(resp, s)
}

func changeSpeed(path string, resp http.ResponseWriter) {
    params := strings.Split(path, "/")
    v, err := strconv.Atoi(params[1])
    if err == nil {
        sendRate = v
    }
    io.WriteString(resp, strconv.Itoa(sendRate))
}
