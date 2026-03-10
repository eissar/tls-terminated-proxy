package main

import (
	"crypto/tls"
	"encoding/base64"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var TargetHost = os.Getenv("PROXY_HOST")

func main() {
	if TargetHost == "" {
		log.Fatal("PROXY_HOST environment variable is required")
	}

	proxyRaw := os.Getenv("PROXY_URL")
	if proxyRaw == "" {
		log.Fatal("PROXY_URL environment variable is required")
	}

	proxyURL, err := url.Parse(proxyRaw)
	if err != nil {
		log.Fatalf("failed to parse PROXY_URL: %v", err)
	}
	if proxyURL.Scheme == "" || proxyURL.Host == "" {
		log.Fatal("PROXY_URL must include scheme and host, e.g. http://proxy.example.com:8080")
	}

	// Prefer explicit env creds; fallback to creds embedded in PROXY_URL.
	username := os.Getenv("PROXY_USERNAME")
	password := os.Getenv("PROXY_PASSWORD")

	if username != "" || password != "" {
		if username == "" || password == "" {
			log.Fatal("both PROXY_USERNAME and PROXY_PASSWORD must be set together")
		}
		proxyURL.User = url.UserPassword(username, password)
	} else if proxyURL.User == nil {
		log.Fatal("proxy auth required: set PROXY_USERNAME/PROXY_PASSWORD or include creds in PROXY_URL")
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.Proxy = http.ProxyURL(proxyURL)

	if proxyURL.User == nil {
		log.Fatal("User was somehow nil")
	}

	user := proxyURL.User.Username()
	pass, _ := proxyURL.User.Password()
	tok := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))

	tr.ProxyConnectHeader = http.Header{
		"Proxy-Authorization": []string{"Basic " + tok},
	}

	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	rp := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = TargetHost
			req.Host = TargetHost
		},
		Transport: tr,
	}

	log.Println("reverse proxy listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", rp))
}
