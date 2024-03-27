package httprouter

import (
	"log"
	"net/http"
	"regexp"
)

type RouterNode struct {
	prefix     string
	handler    http.Handler
	childNodes []*RouterNode
	regex      *regexp.Regexp
}

func createRouterNode(prefix string) *RouterNode {
	// create a compiled regex if the node uses a regex
	var re *regexp.Regexp = nil
	if isRegex(prefix) {
		r, err := regexp.Compile(extractRegex(prefix))
		if err != nil {
			log.Fatalf("Failed to compile regex")
		}
		re = r
	}
	// prefix node wasn't found, then create new, add it to node, and return it
	return &RouterNode{
		prefix:     prefix,
		childNodes: []*RouterNode{},
		handler:    nil,
		regex:      re,
	}
}

func (routerNode *RouterNode) matchChild(prefix string) *RouterNode {
	// iterate through level to find the first matching node for prefix
	for _, child := range routerNode.childNodes {
		// literal matches when equal
		if child.prefix == prefix {
			return child
		}
		// url param always matches
		if child.isURLParam() {
			return child
		}
		// regex matches with a special case
		if child.regex != nil && child.regex.MatchString(prefix) {
			return child
		}
	}
	return nil
}

func (routerNode *RouterNode) findChild(prefix string) *RouterNode {
	// iterate through level to find node for prefix
	for _, child := range routerNode.childNodes {
		if child.prefix == prefix {
			return child
		}
	}
	return nil
}

func (routerNode* RouterNode) insertChild(nodeToInsert *RouterNode) {
	if nodeToInsert.isURLParam() {
		// url param node is added to the end
		routerNode.childNodes = append(routerNode.childNodes, nodeToInsert)
	} else if nodeToInsert.isRegex() {
		// regex node is added before the first url param or regex param
		index := len(routerNode.childNodes)
		for i, child := range routerNode.childNodes {
			if child.isURLParam() || child.isRegex() {
				index = i
				break
			}
		}
		routerNode.childNodes = insert(routerNode.childNodes, index, nodeToInsert)
	} else {
		// literal node is added to the front
		routerNode.childNodes = prepend(routerNode.childNodes, nodeToInsert)
	}
}

func (routerNode *RouterNode) findOrCreateChild(prefix string) *RouterNode {
	// iterate through level to find node with prefix
	node := routerNode.findChild(prefix)
	if node != nil {
		return node
	}
	// prefix node wasn't found, then create new, add it to node, and return it
	node = createRouterNode(prefix)
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
	for _, n := range node.childNodes {
		traverseNode(prefix+"/"+n.prefix, n, routes)
	}
}

func isURLParam(prefix string) bool {
	return prefix[0] == '{' && prefix[len(prefix)-1] == '}'
}

func extractURLParam(prefix string) string {
	return prefix[1:len(prefix)-1]
}

func isRegex(prefix string) bool {
	return prefix[0] == '|' && prefix[len(prefix)-1] == '|'
}

func extractRegex(prefix string) string {
	return prefix[1:len(prefix)-1]
}
