package httprouter

import (
	"hash/fnv"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type tree struct {
	nodes []exactNode
}

func hashPrefix(prefix string) uint64 {
	h := fnv.New64()
	for i := 0; i < len(prefix); i++ {
		h.Write([]byte(prefix))
	}
	return h.Sum64()
}

func (tree *tree) routes() []string {
	routes := make([]string, 0)
	for _, child := range tree.nodes {
		child.next.walkNode(child.next.prefix+" ", &routes)
	}
	return routes
}

func (tree *tree) appendNode(method string) *node {
	node := exactNode{
		hash: hashPrefix(method),
		next: node{
			prefix:     method,
			exactNodes: []exactNode{},
			paramNodes: []paramNode{},
		},
	}
	tree.nodes = append(tree.nodes, node)
	return &tree.nodes[len(tree.nodes)-1].next
}

func (tree *tree) findNode(method string) *node {
	hash := hashPrefix(method)
	for i := range tree.nodes {
		child := &tree.nodes[i]
		if child.hash == hash {
			return &child.next
		}
	}
	return nil
}

func (tree *tree) appendRoute(method string, route string) *node {
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

type node struct {
	prefix     string
	handler    http.Handler
	exactNodes []exactNode // these match when the hash is equal
	paramNodes []paramNode // these match when a regex is matched
}

type exactNode struct {
	hash uint64
	next node
}

type paramNode struct {
	param string
	regex *regexp.Regexp
	next  node
}

func (n *node) matchChild(prefix string) (*node, string) {
	hash := hashPrefix(prefix)
	for i := len(n.exactNodes) - 1; i >= 0; i-- {
		child := n.exactNodes[i]
		if child.hash == hash {
			return &child.next, ""
		}
	}
	for i := len(n.paramNodes) - 1; i >= 0; i-- {
		child := n.paramNodes[i]
		if child.regex == nil {
			return &child.next, child.param
		} else if child.regex.MatchString(prefix) {
			return &child.next, child.param
		}
	}
	return nil, ""
}

func (n *node) appendChild(prefix string) *node {
	newNode := node{
		prefix:     prefix,
		exactNodes: []exactNode{},
		paramNodes: []paramNode{},
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
		paramNode := paramNode{param: param, next: newNode}
		if regStr != "" {
			re, err := regexp.Compile(regStr)
			if err != nil {
				log.Fatalf("Failed to compile regex: %s", regStr)
			}
			paramNode.regex = re
		}
		n.paramNodes = append(n.paramNodes, paramNode)
		return &n.paramNodes[len(n.paramNodes)-1].next
	} else {
		exactNode := exactNode{
			hash: hashPrefix(prefix),
			next: newNode,
		}
		n.exactNodes = append(n.exactNodes, exactNode)
		return &n.exactNodes[len(n.exactNodes)-1].next
	}
}

func (n *node) findOrAppendNode(prefix string) *node {
	for i := range n.exactNodes {
		child := &n.exactNodes[i]
		if child.next.prefix == prefix {
			return &child.next
		}
	}
	for i := range n.paramNodes {
		child := &n.paramNodes[i]
		if child.next.prefix == prefix {
			return &child.next
		}
	}
	child := n.appendChild(prefix)
	return child
}

func (n *node) walkNode(prefix string, routes *[]string) {
	if n.handler != nil {
		*routes = append(*routes, prefix)
	}
	for _, child := range n.exactNodes {
		next := child.next
		next.walkNode(prefix+"/"+next.prefix, routes)
	}
	for _, child := range n.paramNodes {
		next := child.next
		next.walkNode(prefix+"/"+next.prefix, routes)
	}
}
