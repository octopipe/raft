package server

import (
	"net"

	"google.golang.org/grpc"
  raftPb "github.com/octopipe/raft/proto/raft/v1"
)

type Server struct {
  raftPb.UnimplementedRaftServer
  ID string
  address string
  timer int
}

func NewServer(host string, port int) Server {
  newServer := Server{
    ID: "",
    address: "",
    timer: 0,
  }

  return newServer
}

func (s Server) Start() error {
  lis, err := net.Listen("tcp", s.address)
  if err != nil {
    return err
  }
  grpcServer := grpc.NewServer()
  raftPb.RegisterRaftServer(grpcServer, s)
  return grpcServer.Serve(lis)

}

