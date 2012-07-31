package main

import (
	"errors"
	"path"
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
