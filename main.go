package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

//go:embed static
var staticFS embed.FS

func main() {
	port := flag.Int("port", 8000, "port to listen on")
	altPort := flag.Int("alt-port", 0, "secondary port for cross-origin testing (0 = disabled)")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check query params for header configuration
		coop := r.URL.Query().Get("coop")
		coep := r.URL.Query().Get("coep")
		corp := r.URL.Query().Get("corp")

		// Set COOP header
		switch coop {
		case "same-origin":
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		case "same-origin-allow-popups":
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")
		case "unsafe-none":
			w.Header().Set("Cross-Origin-Opener-Policy", "unsafe-none")
		}

		// Set COEP header
		switch coep {
		case "require-corp":
			w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		case "credentialless":
			w.Header().Set("Cross-Origin-Embedder-Policy", "credentialless")
		}

		// Set CORP header (for being embedded)
		switch corp {
		case "same-origin":
			w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		case "same-site":
			w.Header().Set("Cross-Origin-Resource-Policy", "same-site")
		case "cross-origin":
			w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
		}

		// Set Document-Isolation-Policy header (Chrome)
		dip := r.URL.Query().Get("dip")
		switch dip {
		case "isolate-and-require-corp":
			w.Header().Set("Document-Isolation-Policy", "isolate-and-require-corp")
		case "isolate-and-credentialless":
			w.Header().Set("Document-Isolation-Policy", "isolate-and-credentialless")
		}

		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		data, err := staticFS.ReadFile("static" + path)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if strings.HasSuffix(path, ".html") {
			w.Header().Set("Content-Type", "text/html")
		} else if strings.HasSuffix(path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		} else if strings.HasSuffix(path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}

		w.Write(data)
	})

	// Echo endpoint for fetch tests
	http.HandleFunc("/api/echo", func(w http.ResponseWriter, r *http.Request) {
		corp := r.URL.Query().Get("corp")
		switch corp {
		case "same-origin":
			w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		case "same-site":
			w.Header().Set("Cross-Origin-Resource-Policy", "same-site")
		case "cross-origin":
			w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
		}

		w.Header().Set("Content-Type", "application/json")
		
		// Set ACAO header - default to cross-origin.toy.lalitm.com
		acao := r.URL.Query().Get("acao")
		if acao == "" {
			acao = "https://cross-origin.toy.lalitm.com"
		}
		w.Header().Set("Access-Control-Allow-Origin", acao)
		
		// Set Access-Control-Allow-Credentials when ACAO is set
		if acao != "" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		fmt.Fprintf(w, `{"status":"ok","method":"%s","corp":"%s","acao":"%s"}`, r.Method, corp, acao)
	})

	// Proxy endpoint for cross-origin fetch tests
	http.HandleFunc("/api/proxy", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if url == "" {
			http.Error(w, "url parameter required", http.StatusBadRequest)
			return
		}

		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Access-Control-Allow-Origin", "*")
		for k, v := range resp.Header {
			if k != "Access-Control-Allow-Origin" {
				w.Header()[k] = v
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	// Start alt server for cross-origin testing
	if *altPort > 0 {
		go func() {
			altAddr := fmt.Sprintf(":%d", *altPort)
			log.Printf("Alt server (cross-origin) starting on http://localhost%s", altAddr)
			log.Fatal(http.ListenAndServe(altAddr, nil))
		}()
	}

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
