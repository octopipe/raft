package node

import "github.com/octopipe/raft/pkg/server"

type NodeState uint32

const (
  Follower NodeState = iota
  Candidate
  Leader
)

type Node struct {
  server.Server
  currentTerm int
  votedFor int
  log [][]byte
  commitIndex int
  lastApplied int
}

func NewNode() {
}

