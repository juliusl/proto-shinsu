package control

import (
	"fmt"
	"net/http"
)

func CreateNode(transport *NodeTransport) (*Node, error) {
	return &Node{transport: transport}, nil
}

type Node struct {
	transport *NodeTransport
}

func (n *Node) AddAPI(root, term string, streamDesc *StreamDescriptor) {
	n.transport.api[fmt.Sprintf("%s/%s", root, term)] = streamDesc
}

func (n *Node) GetClient() *http.Client {
	return &http.Client{Transport: n.transport}
}
