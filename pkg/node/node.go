package node

import "github.com/octopipe/raft/pkg/server"

type Node struct {
  server.Server
  currentTerm int
  votedFor int
  log [][]byte
  commitIndex int
  lastApplied int
}

func NewNode() {}

