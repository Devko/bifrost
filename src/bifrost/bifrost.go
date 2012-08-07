package main

import (
	"flag"
	conf "github.com/cr0n/goconf"
	"log"
	"net/http"
	"net/http/httputil"
)

var conf_path = *flag.String("c", "config.json", "path to configuration file")
var balance = &balanceMap{data: make(map[string]int)}

const (
	def_addr = ":80"
)

func main() {
	if err := conf.LoadConfig(conf_path); err != nil {
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
	http.HandleFunc("/", proxyHandler)
	if err = http.ListenAndServe(address, nil); err != nil {
		log.Fatalln(err)
	}
}

func proxyHandler(resp http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error :", err)
			http.Error(resp, err.(error).Error(), http.StatusInternalServerError)
		}
	}()
	if p := cleanPath(req.URL.Path); p != req.URL.Path {
		log.Println("cleaned up request url", req.URL.Path, "->", p)
		resp.Header().Set("Location", p)
		resp.WriteHeader(http.StatusMovedPermanently)
		return
	}
	addrs, err := getRoutingTable()
	if err != nil {
		log.Println("failed to read config :", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	dest, err := findRoute(req, addrs)
	if err != nil {
		log.Println("could not find route for request to :", req.URL.Host+req.URL.Path, "error :", err)
		http.NotFound(resp, req)
		return
	}
	httputil.NewSingleHostReverseProxy(dest).ServeHTTP(resp, req)
}
