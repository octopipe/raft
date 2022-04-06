package node

type Server struct {
  ID int
  host string
  port int
}

type Node struct {
  Server
  currentTerm int
  votedFor int
  log [][]byte
  commitIndex int
  lastApplied int
}

type Leader struct {
  Node
  nextIndex []int
  matchIndex []int
}

