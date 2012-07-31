package main

import (
	"flag"
	conf "github.com/AndriyLytvynov/goconf"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var conf_path = *flag.String("c", "config.json", "path to configuration file")

const (
	def_addr = ":80"
)

func main() {
	log.Println("bifrost started.")
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
		log.Println("bifrost stopped.")
	}
}

func proxyHandler(resp http.ResponseWriter, req *http.Request) {
    defer func(){
        if err := recover(); err != nil {
            log.Println("error :", err)
        }
    }()
    log.Println("request to ", req.RequestURI)
    tmp, err := conf.Get("proxy_info")
    if err != nil {
        log.Println("failed to read config :", err)
        http.NotFound(resp, req)
        return
    }
    addrs, ok := tmp.(map[string]interface{})
    if !ok {
        log.Println("config is not formatted well!")
        http.NotFound(resp, req)
        return
    }
    route, err := match(req.URL.Path, addrs)
    if err != nil {
        log.Println("could not find route for request :", req.URL.Path)
        http.NotFound(resp, req)
        return
    }
    dest, err := url.Parse(addrs[route].([]string)[0])
    if err != nil {
        log.Println("config url not valid :", addrs[route].([]string)[0])
        http.NotFound(resp, req)
        return
    }
    httputil.NewSingleHostReverseProxy(dest).ServeHTTP(resp, req)
}
