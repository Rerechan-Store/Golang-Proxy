package main

import (
// Imprt Pakage yang diperlukan
 "fmt"
 "io"
 "log"
 "net"
 "net/http"
 "net/url"
 "strings"
)

const (
// Detail Kongigurasi
 ListeningAddr = "0.0.0.0"
 ListeningPort = "2082"
 DefaultHost   = "127.0.0.1:111"
 Response      = "HTTP/1.1 101 Switching Protocols\r\n\r\n"
)

func main() {
 fmt.Println("\n:-------GoProxy-------:\n")
 fmt.Println("Listening addr: " + ListeningAddr)
 fmt.Println("Listening port: " + ListeningPort + "\n")
 fmt.Println(":---------------------:\n")

 http.HandleFunc("/", handleRequest)
 log.Fatal(http.ListenAndServe(ListeningAddr+":"+ListeningPort, nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
 hostPort := r.Header.Get("X-Real-Host")
 if hostPort == "" {
  hostPort = DefaultHost
 }

 split := r.Header.Get("X-Split")
 if split != "" {
  io.CopyN(io.Discard, r.Body, 4096)
 }

 if hostPort != "" {
  passwd := r.Header.Get("X-Pass")
  if passwd == "" {
   passwd = r.FormValue("X-Pass")
  }

  if passwd == "" {
   passwd = r.URL.Query().Get("X-Pass")
  }

  if passwd == "" {
   passwd = r.PostFormValue("X-Pass")
  }

  if passwd == "" {
   passwd = r.Referer()
  }

  if passwd == "" {
   passwd = r.UserAgent()
  }

  if passwd != "" && passwd != PASS {
   w.WriteHeader(http.StatusBadRequest)
   return
  }

  if hostPort == DefaultHost || strings.HasPrefix(hostPort, "127.0.0.1") || strings.HasPrefix(hostPort, "localhost") {
   proxyRequest(w, r, hostPort)
  } else {
   w.WriteHeader(http.StatusForbidden)
  }
 } else {
  log.Println("- No X-Real-Host!")
  w.WriteHeader(http.StatusBadRequest)
 }
}

func proxyRequest(w http.ResponseWriter, r *http.Request, hostPort string) {
 targetURL := "http://" + hostPort + r.RequestURI
 target, err := url.Parse(targetURL)
 if err != nil {
  log.Println("Failed to parse target URL:", err)
  w.WriteHeader(http.StatusInternalServerError)
  return
 }

 proxy := &httputil.ReverseProxy{
  Director: func(req *http.Request) {
   req.URL.Scheme = target.Scheme
   req.URL.Host = target.Host
   req.Host = target.Host
  },
 }

 proxy.ServeHTTP(w, r)
}
