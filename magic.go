package main

import (
	"errors"
	"fmt"
	conf "github.com/cr0n/goconf"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
)

func match(pat string, addrs map[string]interface{}) (string, error) {
	res := ""
	for k, _ := range addrs {
		n := len(k)
		if n == 0 || (k[n-1] != '/' && k != pat) || len(pat) < n || pat[0:n] != k {
			continue
		}
		if len(k) > len(res) {
			res = k
		}
	}
	if res == "" {
		return res, errors.New("no match found")
	}
	return res, nil
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

type balanceMap struct {
	sync.Mutex
	data map[string]int
}

func (m *balanceMap) getNext(key string, routes []interface{}) (string, error) {
	m.Lock()
	defer m.Unlock()
	cur, ok := m.data[key]
	if !ok {
		cur = 0
	} else {
		cur = (cur + 1) % len(routes)
	}
	m.data[key] = cur
	res, ok := routes[cur].(string)
	if !ok {
		return "", errors.New("bad configuration : " + fmt.Sprint(routes[0]) + " should be a string")
	}
	return res, nil
}

func getRoutingTable() (map[string]interface{}, error) {
	tmp, err := conf.Get("proxy_table")
	if err != nil {
		return nil, err
	}
	addrs, ok := tmp.(map[string]interface{})
	if !ok {
		return nil, errors.New("bad configuration : proxy_table should be map[string][]string")
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
	for i := 0; i < len(routes); i++ {
		dest, err := balance.getNext(route, routes)
		if err != nil {
			return nil, err
		}
		res, err := url.Parse(dest)
		if err != nil {
			log.Println("invalid URL", dest, "for pattern", route, "error:", err)
			continue
		}
		return res, nil
	}
	return nil, errors.New("all urls for pattern " + route + " are invalid, abort routing")
}

func joinUrls(a, b string) string {
	aslash, bslash := strings.HasSuffix(a, "/"), strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
