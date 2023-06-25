package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	viper "github.com/spf13/viper"
	logrus "gopkg.in/sirupsen/logrus.v1"
)

type DataResponse struct {
	Hostname    string      `json:"hostname,omitempty"`
	Platform    string      `json:"platform,omitempty"`
	IP          []string    `json:"ip,omitempty"`
	Headers     http.Header `json:"header,omitempty"`
	Environment []string    `json:"env,omitempty"`
}

var port string

func init() {
	flag.StringVar(&port, "port", "8080", "give me a port number")

	lvl, err := logrus.ParseLevel(viper.GetString("loglevel"))
	if err != nil {
		lvl = logrus.WarnLevel
	}
	logrus.SetLevel(lvl)
}

func main() {
	flag.Parse()

	http.HandleFunc("/", index)
	http.HandleFunc("/api", api)

	log.Println("Starting up on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		http.ListenAndServe(":"+port, nil)
	}
}

func index(w http.ResponseWriter, req *http.Request) {
	u, _ := url.Parse(req.URL.String())
	queryParams := u.Query()

	wait := queryParams.Get("wait")
	if len(wait) > 0 {
		duration, err := time.ParseDuration(wait)
		if err == nil {
			time.Sleep(duration)
		}
	}

	data := fetchData(req)
	fmt.Fprintf(os.Stdout, "I'm %s\n", data.Hostname)
	fmt.Fprintf(w, "Hello again, I'm %s running on %s\n\n", data.Hostname, data.Platform)

	for _, ip := range data.IP {
		fmt.Fprintln(w, "IP:", ip)
	}

	for _, env := range data.Environment {
		fmt.Fprintln(w, "ENV:", env)
	}
	req.Write(w)
}

func api(w http.ResponseWriter, req *http.Request) {
	data := fetchData(req)
	json.NewEncoder(w).Encode(data)
}

func fetchData(req *http.Request) DataResponse {
	hostname, _ := os.Hostname()
	data := DataResponse{
		hostname,
		fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		[]string{},
		req.Header,
		os.Environ(),
	}

	ifaces, _ := net.Interfaces()
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			data.IP = append(data.IP, ip.String())
		}
	}

	return data
}
