package main

import (
  "fmt"
  "os"
  "net/http"
  "net/http/httputil"
  "net/url"
)

type Server interface {
  Adress() string
  ISAlive() bool
  Serve(rw http.ResponseWriter, req *http.Request)
}

type simpleServer struct {
  adress string
  proxy *httputil.ReverseProxy

}

func handleError(error error,message string) {
if error != nil {
  fmt.Println("[ERROR]:",message, error)
  os.Exit(1)
}
}


func createServer(adress string) *simpleServer {
  serverUrl,error := url.Parse(adress)
  handleError(error,"failed to parse url")
return &simpleServer{
  adress: adress,
  proxy: httputil.NewSingleHostReverseProxy(serverUrl),
}

}

func (s *simpleServer) Adress() string {
  return s.adress
}

func (s *simpleServer) ISAlive() bool {
  return true
}

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
  s.proxy.ServeHTTP(rw,req)
}


type LoadBalancer struct {
  port string
  rbCount int
  servers []Server
}


func createLoadBalancer(port string ,servers []Server) *LoadBalancer {
  return &LoadBalancer{
    port: port,
    rbCount:0,
    servers: servers,
  }
}


func (lb *LoadBalancer) GetAvailableServer() Server {
  server := lb.servers[lb.rbCount%len(lb.servers)]

  for !server.ISAlive() {
    lb.rbCount++
    server = lb.servers[lb.rbCount%len(lb.servers)]
}
  lb.rbCount++
  return server
}
func (lb *LoadBalancer) serveProxy(rw http.ResponseWriter,req *http.Request) {
  targetServer := lb.GetAvailableServer()
  fmt.Println("[SUCCESS]: Forwarding request to ",targetServer.Adress())
  targetServer.Serve(rw,req)
}

func main(){
  servers := []Server {
    createServer("https://instagram.com"),
    createServer("https://facebook.com"),
    createServer("https://github.com"),
    createServer("https://youtube.com"),
}

  lb := createLoadBalancer("8080",servers)

  redirect := func(rw http.ResponseWriter, req *http.Request) {
    fmt.Println("[INFO]: Received Request")
    lb.serveProxy(rw,req)
  }
  http.HandleFunc("/", redirect)

  fmt.Println("[INFO]: Starting server on port", lb.port)
  http.ListenAndServe(":"+lb.port,nil)
}

