package server

type Server struct {
  ID string
  host string
  port int
  timer int
}

func NewServer(host string, port int) Server {
  newServer := Server{
    ID: "",
    host: host,
    port: port,
    timer: 0,
  }

  return newServer
}
