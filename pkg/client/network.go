package client

import (
	networkPb "github.com/octopipe/raft/proto/network/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewNetworkClient(target string) (networkPb.NetworkClient, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := networkPb.NewNetworkClient(conn)
	return client, nil
}
