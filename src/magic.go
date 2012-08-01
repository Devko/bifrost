package main

import (
	"errors"
	"fmt"
	"path"
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
