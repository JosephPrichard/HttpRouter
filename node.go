package httprouter

import (
	"log"
	"net/http"
	"regexp"
)

type RouterNode struct {
	prefix   string
	handler  http.Handler
	children []*RouterNode
	regex    *regexp.Regexp
}

func createRouterNode(prefix string) *RouterNode {
	var re *regexp.Regexp = nil
	if isRegex(prefix) {
		r, err := regexp.Compile(extractRegex(prefix))
		if err != nil {
			log.Fatalf("Failed to compile regex")
		}
		re = r
	}
	return &RouterNode{
		prefix:   prefix,
		children: []*RouterNode{},
		handler:  nil,
		regex:    re,
	}
}

func (routerNode *RouterNode) matchChild(prefix string) *RouterNode {
	for _, child := range routerNode.children {
		if child.prefix == prefix {
			return child
		}
		if child.isURLParam() {
			return child
		}
		if child.regex != nil && child.regex.MatchString(prefix) {
			return child
		}
	}
	return nil
}

func insertNode(slice []*RouterNode, index int, value *RouterNode) []*RouterNode {
	if len(slice) < 1 {
		return []*RouterNode{value}
	}
	slice = append(slice[:index+1], slice[index:]...)
	slice[index] = value
	return slice
}

func prependNode(slice []*RouterNode, value *RouterNode) []*RouterNode {
	return append([]*RouterNode{value}, slice...)
}

func (routerNode *RouterNode) insertChild(node *RouterNode) {
	if node.isURLParam() {
		routerNode.children = append(routerNode.children, node)
	} else if node.isRegex() {
		index := len(routerNode.children)
		for i, child := range routerNode.children {
			if child.isURLParam() || child.isRegex() {
				index = i
				break
			}
		}
		routerNode.children = insertNode(routerNode.children, index, node)
	} else {
		routerNode.children = prependNode(routerNode.children, node)
	}
}

func (routerNode *RouterNode) findOrCreateChild(prefix string) *RouterNode {
	for _, child := range routerNode.children {
		if child.prefix == prefix {
			return child
		}
	}
	node := createRouterNode(prefix)
	routerNode.insertChild(node)
	return node
}

func (routerNode *RouterNode) isRegex() bool {
	return isRegex(routerNode.prefix)
}

func (routerNode *RouterNode) isURLParam() bool {
	return isURLParam(routerNode.prefix)
}

func (routerNode *RouterNode) getURLParam() (bool, string) {
	return routerNode.isURLParam(), extractURLParam(routerNode.prefix)
}

func traverseNode(prefix string, node *RouterNode, routes *[]string) {
	if node.handler != nil {
		*routes = append(*routes, prefix)
	}
	for _, n := range node.children {
		traverseNode(prefix+"/"+n.prefix, n, routes)
	}
}

func isURLParam(prefix string) bool {
	return prefix[0] == '{' && prefix[len(prefix)-1] == '}'
}

func extractURLParam(prefix string) string {
	return prefix[1 : len(prefix)-1]
}

func isRegex(prefix string) bool {
	return prefix[0] == '[' && prefix[len(prefix)-1] == ']'
}

func extractRegex(prefix string) string {
	return prefix[1 : len(prefix)-1]
}
