package node

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/octopipe/raft/pkg/client"
	raftPb "github.com/octopipe/raft/proto/raft/v1"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
	"time"

	networkPb "github.com/octopipe/raft/proto/network/v1"
)

type NodeState uint32

const (
	Follower NodeState = iota
	Candidate
	Leader
)

type Node struct {
	networkPb.NetworkServer
	raftPb.RaftServer

	ID    uuid.UUID
	host  string
	port  int
	group []string

	currentTerm int
	votedFor    []string
	log         [][]byte
	commitIndex int
	lastApplied int

	timeout time.Duration
	status  NodeState
}

func NewNode(host string, port int, target string) (Node, error) {
	r := rand.Intn(10)

	n := Node{
		ID:      uuid.New(),
		status:  Follower,
		timeout: time.Duration(r) * time.Millisecond,
		host:    host,
		port:    port,
		group:   []string{},
	}

	if target != "" {
		networkClient, err := client.NewNetworkClient(target)
		if err != nil {
			return Node{}, err
		}

		request := &networkPb.JoinRequest{Host: []byte(target)}
		reply, err := networkClient.Join(context.Background(), request)
		if err != nil {
			log.Fatalln(err)
		}

		for _, h := range reply.Hosts {
			n.group = append(n.group, string(h))
		}
	}

	go n.startHeartbeat()

	return n, nil
}

func (n Node) StartServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", n.host, n.port))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	raftPb.RegisterRaftServer(grpcServer, n)
	networkPb.RegisterNetworkServer(grpcServer, n)
	err = grpcServer.Serve(lis)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) startElectionTimer() {
	select {
	case <-time.After(n.timeout):
		n.status = Candidate
		n.currentTerm += 1
		n.votedFor = append(n.votedFor, n.ID.String())
		request := &raftPb.RequestVoteRequest{
			Term:        int64(n.currentTerm),
			CandidateId: []byte(n.ID.String()),
		}

		for _, g := range n.group {
			raftClient, err := client.NewRaftClient(g)
			if err != nil {
				log.Fatalln(err)
			}
			_, err = raftClient.RequestVote(context.Background(), request)
			if err != nil {
				log.Fatalln(err)
			}
		}

	}
}

func (n *Node) doLeaderRules() {
	for _, g := range n.group {
		raftClient, err := client.NewRaftClient(g)
		if err != nil {
			log.Fatalln(err)
		}
		request := &raftPb.AppendEntriesRequest{
			Term:     int64(n.currentTerm),
			LeaderId: []byte(n.ID.String()),
			//TODO: COMMIT INDEX RULES
		}
		_, err = raftClient.AppendEntries(context.Background(), request)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func (n *Node) doCandidateRules() {
	// Increment currentTerm
	n.currentTerm += 1

	// Vote for self
	n.votedFor = append(n.votedFor, n.ID.String())

	// Send request vote to all nodes
	request := &raftPb.RequestVoteRequest{
		Term:        int64(n.currentTerm),
		CandidateId: []byte(n.ID.String()),
	}

	for _, g := range n.group {
		raftClient, err := client.NewRaftClient(g)
		if err != nil {
			log.Fatalln(err)
		}
		reply, err := raftClient.RequestVote(context.Background(), request)
		if err != nil {
			log.Fatalln(err)
		}

		if reply.VoteGranted {
			n.votedFor = append(n.votedFor, g)
		}

		// If votes received from majority of servers: become leader
		if len(n.votedFor) > (len(n.group) / 2) {
			n.status = Leader
			// Upon election: send initial empty AppendEntries RPCs
			n.doLeaderRules()
		}
	}
}

func (n *Node) startHeartbeat() {
	t := time.NewTicker(n.timeout)

	// TODO: If election timeout elapses: start new election

	select {
	case <-t.C:
		switch n.status {
		case Follower:
			n.status = Candidate
			n.doCandidateRules()
			break
		case Candidate:
			n.doCandidateRules()
			break
		case Leader:
			n.doLeaderRules()
			break
		default:
			log.Fatalln("Unrecognized status!")
		}

	}
}

func (n *Node) AddToGroup(host string) {
	n.group = append(n.group, host)
}

func (n *Node) setVotedFor(cadidateId string) {
	n.votedFor = append(n.votedFor, cadidateId)
}

func (n *Node) setState(state NodeState) {
	n.status = state
}

func (n Node) Join(ctx context.Context, request *networkPb.JoinRequest) (*networkPb.JoinReply, error) {
	log.Println("Received join node rpc")
	hosts := [][]byte{}

	n.AddToGroup(string(request.Host))
	for _, g := range n.group {
		hosts = append(hosts, []byte(g))
	}

	reply := &networkPb.JoinReply{Hosts: hosts}

	return reply, nil
}

func (n Node) RequestVote(ctx context.Context, request *raftPb.RequestVoteRequest) (*raftPb.RequestVoteReply, error) {
	log.Println("Received request vote...")

	voteGranted := true
	if request.Term < int64(n.currentTerm) {
		voteGranted = false
	}

	reply := &raftPb.RequestVoteReply{
		Term:        int64(n.currentTerm),
		VoteGranted: voteGranted,
	}

	return reply, nil
}

func (n Node) AppendEntries(ctx context.Context, request *raftPb.AppendEntriesRequest) (*raftPb.AppendEntriesReply, error) {
	log.Println("Received append entry...")

	reply := &raftPb.AppendEntriesReply{
		Term:    int64(n.currentTerm),
		Success: true,
	}

	// If AppendEntries RPC received from new leader: convert to follower
	n.setState(Follower)

	return reply, nil
}
