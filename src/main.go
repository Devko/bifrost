package main

import (
	"errors"
	"flag"
	"fmt"
	conf "github.com/AndriyLytvynov/goconf"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var conf_path = *flag.String("c", "config.json", "path to configuration file")
var balance = &balanceMap{data: make(map[string]int)}

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
	}
}

func proxyHandler(resp http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("error :", err)
			http.Error(resp, err.(error).Error(), http.StatusInternalServerError)
		}
	}()
	log.Println("request to ", req.Host+req.URL.Path)
	if p := cleanPath(req.URL.Path); p != req.URL.Path {
		log.Println("cleaned up request url", req.URL.Path, "->", p)
		resp.Header().Set("Location", p)
		resp.WriteHeader(http.StatusMovedPermanently)
		return
	}
	addrs, err := getConf()
	if err != nil {
		log.Println("failed to read config :", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	dest, err := findRoute(req, addrs)
	if err != nil {
		log.Println("could not find route for request :", err)
		http.NotFound(resp, req)
		return
	}
	httputil.NewSingleHostReverseProxy(dest).ServeHTTP(resp, req)
}

func getConf() (map[string]interface{}, error) {
	tmp, err := conf.Get("proxy_info")
	if err != nil {
		return nil, err
	}
	addrs, ok := tmp.(map[string]interface{})
	if !ok {
		return nil, errors.New("bad configuration : proxy_info should be map[string][]string")
	}
	return addrs, nil
}

func findRoute(req *http.Request, addrs map[string]interface{}) (*url.URL, error) {
	route, err := match(req.URL.Path, addrs)
	if err != nil || addrs[route] == nil {
		route, err = match(req.Host+req.URL.Path, addrs)
		if err != nil || addrs[route] == nil {
			return nil, err
		}
	}
	routes, ok := addrs[route].([]interface{})
	if !ok {
		return nil, errors.New("bad configuration : " + fmt.Sprint(addrs[route]) + " should be []string")
	}
	dest, err := balance.getNext(route, routes)
	if err != nil {
		return nil, err
	}
	res, err := url.Parse(dest)
	if err != nil {
		return nil, err
	}
	return res, nil
}
