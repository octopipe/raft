package leader

import "github.com/octopipe/raft/pkg/node"

type Leader struct {
  node.Node
  nextIndex []int
  matchIndex []int
}

func NewLeader() {
}
