package main

import (
	"flag"
	conf "github.com/cr0n/goconf"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"fmt"
	"strings"
)

var conf_path = flag.String("c", "config.json", "path to configuration file")
var balance = &balanceMap{data: make(map[string]int)}

const (
	def_addr = ":8080"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	flag.Parse()
	if err := conf.LoadConfig(*conf_path); err != nil {
		log.Fatal("could not read configuration file :", err)
	}
	addr, err := conf.Get("addr")
	if err != nil {
		log.Println("addr parameter not specified, using default -", def_addr)
		addr = def_addr
	}
	address, ok := addr.(string)
	if !ok {
		log.Println("invalid addr parameter specified, using default -", def_addr)
		address = def_addr
	}
	http.Handle("/challenges/", &httputil.ReverseProxy{Director: proxyDirector})
	http.HandleFunc("/", mainHandler)
	http.Handle("/notfound", http.NotFoundHandler())
	if err = http.ListenAndServe(address, nil); err != nil {
		log.Fatalln(err)
	}
}

func proxyDirector(req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error :", err)
			req.URL = nil
		}
	}()
	addrs, err := getRoutingTable()
	if err != nil {
		log.Println("failed to read config :", err)
		req.URL = nil
		return
	}
	dest, err := findRoute(req, addrs)
	if err != nil {
		log.Println("could not find route for request to :", req.URL.Host+req.URL.Path, "error :", err)
		if req.URL, err = url.Parse("http://localhost/notfound"); err != nil {
			log.Println("failed to route to 404 :", err)
			req.URL = nil
		}
		return
	}
	req.URL.Scheme = dest.Scheme
	req.URL.Host = dest.Host
	req.URL.Path = joinUrls(dest.Path, req.URL.Path)
	
	
	log.Println("dest.Path: ", dest.Path)
	log.Println("req.URL.Path: ", req.URL.Path)
	req.URL.Path = strings.Replace(req.URL.Path, "/challenges", "", 1)
	log.Println("req.URL.Path: ", req.URL.Path)
	
	if dest.RawQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = dest.RawQuery + req.URL.RawQuery
		log.Println("RawQuery: ", req.URL.RawQuery)
	} else {
		req.URL.RawQuery = dest.RawQuery + "&" + req.URL.RawQuery
		log.Println("RawQuery: ", req.URL.RawQuery)
	}
}
