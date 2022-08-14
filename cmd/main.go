package main

import (
	"github.com/octopipe/raft/pkg/node"
	"log"
	"sync"
)

type Node struct {
	host   string
	port   int
	target string
}

func main() {
	nodes := []Node{
		{host: "0.0.0.0", port: 8081, target: ""},
		{host: "0.0.0.0", port: 8082, target: "0.0.0.0:8081"},
		{host: "0.0.0.0", port: 8083, target: "0.0.0.0:8081"},
	}

	var wg sync.WaitGroup
	for _, n := range nodes {
		wg.Add(1)
		go func(n Node) {
			defer wg.Done()

			newNode, err := node.NewNode(n.host, n.port, n.target)
			log.Println("Starting server...")
			err = newNode.StartServer()
			if err != nil {
				log.Fatalln(err.Error())
			}
		}(n)
	}

	wg.Wait()

}
