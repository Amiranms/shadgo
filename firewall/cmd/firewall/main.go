//go:build !solution

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	_ "log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/exec"
	"strings"
)

func getHostPort(addr string) (host string, port string, err error) {
	addrHostPort := strings.TrimLeft(addr, "http://")
	addrHostPort = strings.TrimLeft(addrHostPort, "https://")
	errInvalidAddr := errors.New("invalid addr")
	if !strings.Contains(addrHostPort, ":") {
		return "", "", errInvalidAddr
	}
	parts := strings.Split(addrHostPort, ":")

	if len(parts) < 2 {
		return "", "", errInvalidAddr
	}
	return parts[0], parts[1], nil
}

func runService(path, port string) error {
	cmd := exec.Command("go", "run", path, "-port", port)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("service starting errored: %w", err)
	}
	return nil
}

const (
	ServiceLocation = "cmd/service/main.go"
)

// func forwardPass(w http.ResponseWriter, r *http.Request) {
// 	b, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		fmt.Printf("read from body errored: %v", err)
// 	}
// 	br := bytes.NewReader(b)

// 	targetUrl := "http://localhost:8899" + r.URL.Path
// 	if r.URL.RawQuery != "" {
// 		targetUrl += "?" + r.URL.RawQuery
// 	}

// 	req, err := http.NewRequest(r.Method, targetUrl, br)

// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	req.Header = r.Header.Clone()

// 	client := &http.Client{}

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadGateway)
// 		return
// 	}

// 	defer resp.Body.Close()

// 	for k, v := range resp.Header {
// 		for _, vv := range v {
// 			w.Header().Add(k, vv)
// 		}
// 	}

// 	w.WriteHeader(resp.StatusCode)

// 	_, err = io.Copy(w, resp.Body)

// 	if err != nil {
// 		http.Error(w, "Core service error", http.StatusBadGateway)
// 		return
// 	}
// }

func parseFlags(path, saddr, addr *string) {
	flag.StringVar(path, "conf", "configs/example.yaml", "config with rules")
	flag.StringVar(saddr, "service-addr", "http://localhost:8811", "addres where the server will be launched")
	flag.StringVar(addr, "addr", "http://localhost:8810", "firewall address")
	flag.Parse()
}

func main() {
	var confPath, saddr, addr string
	parseFlags(&confPath, &saddr, &addr)

	_, err := url.Parse(addr)

	if err != nil {
		panic("invalid firewall addres")
	}
	_, port, _ := getHostPort(addr)

	_, err = url.Parse(saddr)

	if err != nil {
		panic("invalid service addres")
	}

	_, sport, _ := getHostPort(saddr)

	err = runService(ServiceLocation, sport)

	if err != nil {
		panic("service launch failed")
	}
	target, _ := url.Parse(saddr)
	proxy := httputil.NewSingleHostReverseProxy(target)

	fr := NewYAMLFileReader(confPath)
	RulesExec := NewRulesYaml(fr)
	err = RulesExec.ParseRules()
	if err != nil {
		panic(err)
	}
	RulesExec.CompileRules()

	f := NewFirewall(RulesExec)

	http.HandleFunc("/", f.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})))

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
