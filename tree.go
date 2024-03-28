package httprouter

import (
	"hash/fnv"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type Tree struct {
	nodes []ExactNode
}

func hashPrefix(prefix string) uint64 {
	h := fnv.New64()
	for i := 0; i < len(prefix); i++ {
		h.Write([]byte(prefix))
	}
	return h.Sum64()
}

func (tree *Tree) routes() []string {
	routes := make([]string, 0)
	for _, child := range tree.nodes {
		child.next.walkNode(child.next.prefix+" ", &routes)
	}
	return routes
}

func (tree *Tree) appendNode(method string) *Node {
	node := ExactNode{
		hash: hashPrefix(method),
		next: Node{
			prefix:     method,
			exactNodes: []ExactNode{},
			paramNodes: []ParamNode{},
		},
	}
	tree.nodes = append(tree.nodes, node)
	return &tree.nodes[len(tree.nodes)-1].next
}

func (tree *Tree) findNode(method string) *Node {
	hash := hashPrefix(method)
	for i := range tree.nodes {
		child := &tree.nodes[i]
		if child.hash == hash {
			return &child.next
		}
	}
	return nil
}

func (tree *Tree) appendRoute(method string, route string) *Node {
	node := tree.findNode(method)
	if node == nil {
		node = tree.appendNode(method)
	}

	for _, prefix := range strings.Split(route, "/") {
		if prefix == "" {
			continue
		}
		child := node.findOrAppendNode(prefix)
		node = child
	}
	return node
}

type Node struct {
	prefix     string
	handler    http.Handler
	exactNodes []ExactNode // these match when the hash is equal
	paramNodes []ParamNode // these match when a regex is matched
}

type ExactNode struct {
	hash uint64
	next Node
}

type ParamNode struct {
	param string
	regex *regexp.Regexp
	next  Node
}

func (node *Node) matchChild(prefix string) (*Node, string) {
	hash := hashPrefix(prefix)
	for i := len(node.exactNodes) - 1; i >= 0; i-- {
		child := node.exactNodes[i]
		if child.hash == hash {
			return &child.next, ""
		}
	}
	for i := len(node.paramNodes) - 1; i >= 0; i-- {
		child := node.paramNodes[i]
		if child.regex == nil {
			return &child.next, child.param
		} else if child.regex.MatchString(prefix) {
			return &child.next, child.param
		}
	}
	return nil, ""
}

func (node *Node) appendChild(prefix string) *Node {
	newNode := Node{
		prefix:     prefix,
		exactNodes: []ExactNode{},
		paramNodes: []ParamNode{},
	}

	var param, regStr string
	isParam := prefix[0] == '{' && prefix[len(prefix)-1] == '}'
	if isParam {
		regIndex := strings.Index(prefix, ":")
		if regIndex < 0 {
			param = prefix[1 : len(prefix)-1]
			regStr = ""
		} else {
			param = prefix[1:regIndex]
			regStr = prefix[regIndex+1 : len(prefix)-1]
		}
	}

	if param != "" {
		paramNode := ParamNode{param: param, next: newNode}
		if regStr != "" {
			re, err := regexp.Compile(regStr)
			if err != nil {
				log.Fatalf("Failed to compile regex: %s", regStr)
			}
			paramNode.regex = re
		}
		node.paramNodes = append(node.paramNodes, paramNode)
		return &node.paramNodes[len(node.paramNodes)-1].next
	} else {
		exactNode := ExactNode{
			hash: hashPrefix(prefix),
			next: newNode,
		}
		node.exactNodes = append(node.exactNodes, exactNode)
		return &node.exactNodes[len(node.exactNodes)-1].next
	}
}

func (node *Node) findOrAppendNode(prefix string) *Node {
	for i := range node.exactNodes {
		child := &node.exactNodes[i]
		if child.next.prefix == prefix {
			return &child.next
		}
	}
	for i := range node.paramNodes {
		child := &node.paramNodes[i]
		if child.next.prefix == prefix {
			return &child.next
		}
	}
	child := node.appendChild(prefix)
	return child
}

func (node *Node) walkNode(prefix string, routes *[]string) {
	if node.handler != nil {
		*routes = append(*routes, prefix)
	}
	for _, child := range node.exactNodes {
		next := child.next
		next.walkNode(prefix+"/"+next.prefix, routes)
	}
	for _, child := range node.paramNodes {
		next := child.next
		next.walkNode(prefix+"/"+next.prefix, routes)
	}
}
