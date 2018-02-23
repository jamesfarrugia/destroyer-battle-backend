package main

import (
	"flag"
	"net/http"
	"os"

	logging "github.com/op/go-logging"
)

// Logger
var log = logging.MustGetLogger("socks5-app")

// App address
var addr = flag.String("addr", ":8080", "http service address")

func main() {
	logFormat := logging.MustStringFormatter(
		"%{color}%{time:15:04:05.000} %{shortfunc} ▶ \t%{level:.4s} %{id:03x}%{color:reset} %{message}",
	)
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logBackendFormatter := logging.NewBackendFormatter(logBackend, logFormat)
	logging.SetBackend(logBackend, logBackendFormatter)
	log.Info("Golang Destroyer Game! - James Farrugia 2018")

	clients = make(map[int]*Client)

	http.HandleFunc("/", httpIndex)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func httpIndex(w http.ResponseWriter, r *http.Request) {
	log.Info(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	http.ServeFile(w, r, "index.html")
}
