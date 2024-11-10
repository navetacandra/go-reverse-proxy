package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

var proxyConf [][]string

func generateRSA() {
	cmd := exec.Command("openssl", "req", "-x509", "-newkey", "rsa:4096", "-keyout", "./key.pem", "-out", "./cert.pem", "-days", "365", "-nodes", "-subj", "/C=US/ST=California/L=San Francisco/O=My Org/OU=Org Unit/CN=localhost") // Run the command
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Command execution failed: %v", err)
	}
}

func readConfig() {
	file, err := os.ReadFile("./proxy.conf")
	if err != nil {
		panic(err)
	}

	content := string(file)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		splitted := strings.Split(line, " ")
		if len(splitted) == 2 {
			proxyConf = append(proxyConf, splitted)
		}
	}
}

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostname := strings.Split(r.Host, ":")[0]
		var matchProxy []string
		for _, proxy := range proxyConf {
			if proxy[0] == hostname {
				matchProxy = proxy
				break
			}
		}

		if len(matchProxy) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		target, err := url.Parse(matchProxy[1])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		proxy.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Disable SSL certificate validation
			},
		}

		r.URL.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = target.Host
		proxy.ServeHTTP(w, r)
	}
}

func main() {
	generateRSA()
	readConfig()
	http.HandleFunc("/", handler())

	log.Println("Starting reverse proxy on :443")
	err := http.ListenAndServeTLS(":443", "./cert.pem", "./key.pem", nil)
	if err != nil {
		log.Fatal(err)
	}
}
