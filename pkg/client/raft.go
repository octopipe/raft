package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	raftPb "github.com/octopipe/raft/proto/raft/v1"
)

func NewRaftClient(target string) (raftPb.RaftClient, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := raftPb.NewRaftClient(conn)
	return client, nil
}
