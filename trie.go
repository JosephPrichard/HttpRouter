package httprouter

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Trie[v any] struct {
	roots      map[string]*[]Node[v]
	regexCache map[string]*regexp.Regexp
}

type Node[v any] struct {
	path     string
	value    *v
	children []Node[v]
}

func newTrie[v any]() Trie[v] {
	return Trie[v]{
		roots:      make(map[string]*[]Node[v]),
		regexCache: make(map[string]*regexp.Regexp),
	}
}

func newNode[v any](path string, value v) Node[v] {
	return Node[v]{
		path:     path,
		children: make([]Node[v], 0),
		value:    &value,
	}
}

func (trie *Trie[v]) findRoot(method string) *[]Node[v] {
	nodes, ok := trie.roots[method]
	if !ok {
		nodes := make([]Node[v], 0)
		trie.roots[method] = &nodes
		return &nodes
	}
	return nodes
}

func (trie *Trie[v]) insert(method string, path string, value v) {
	nodes := trie.findRoot(method)

	keepSearching := true
	pathIndex := 0
	for keepSearching {
		keepSearching = false
		for i := range *nodes {
			curr := &(*nodes)[i]

			p := 0
			for (pathIndex+p) < len(path) && p < len(curr.path) {
				if path[pathIndex+p] != curr.path[p] {
					break
				}
				p += 1
			}

			if p != 0 {
				if pathIndex+p == len(path) && p == len(curr.path) {
					// case 1: ins path is the same as the curr path - just set the value
					curr.value = &value
					return
				} else if pathIndex+p == len(path) && p < len(curr.path) {
					// case 2: ins path fits inside the curr path - split at where ins path ends
					curr.split(p)
					curr.value = &value
					return
				} else if pathIndex+p < len(path) && p == len(curr.path) {
					// case 3: curr path fits inside the ins path - traverse curr node's children
					nodes = &curr.children
					keepSearching = true
					pathIndex += p
					break
				} else if pathIndex+p < len(path) && p < len(curr.path) {
					// case 4: neither path reaches the end - split and traverse curr node's children
					curr.split(p)
					curr.value = nil
					nodes = &curr.children
					keepSearching = true
					pathIndex += p
					break
				} else {
					// unkown case
					panic("Unknown case for inserting node in radix trie: this is a bug")
				}
			}
		}
	}

	*nodes = append(*nodes, newNode[v](path[pathIndex:], value))
}

func (node *Node[v]) split(splitIndex int) {
	if node.path[:splitIndex] == "" {
		return
	}
	next := Node[v]{
		path:     node.path[splitIndex:],
		value:    node.value,
		children: node.children,
	}
	node.value = nil
	node.path = node.path[:splitIndex]
	node.children = []Node[v]{next}
}

func debugNodes[v any](nodes []Node[v], indent int) string {
	str := ""
	for i := range nodes {
		n := &nodes[i]

		str += "\n"
		for i := 0; i < indent*4; i++ {
			str += " "
		}

		valStr := ""
		if n.value == nil {
			valStr = "nil"
		} else {
			valStr = fmt.Sprintf("%v", *n.value)
		}
		str += fmt.Sprintf("%s -> %v", n.path, valStr)

		str += debugNodes(n.children, indent+1)
	}
	return str
}

func (trie *Trie[v]) getRegex(regexStr string) (*regexp.Regexp, error) {
	var re *regexp.Regexp
	if regexStr != "" {
		rexp, ok := trie.regexCache[regexStr]
		if !ok {
			rexp, err := regexp.Compile(regexStr)
			if err != nil {
				return nil, fmt.Errorf("failed to compile regex %s", regexStr)
			}
			trie.regexCache[regexStr] = rexp
			re = rexp
		} else {
			re = rexp
		}
	}
	return re, nil
}

func (trie *Trie[v]) matchNode(nodeIdx int, val string, path string) (bool, int, string, error) {
	const (
		ParamDelim = ':'
		RegexDelim = '$'
	)

	var name strings.Builder
	var regex strings.Builder

	var sb = &name

	for nodeIdx < len(path) {
		if path[nodeIdx] == '/' {
			break
		} else if path[nodeIdx] == RegexDelim && sb == &name {
			sb = &regex
		} else {
			sb.WriteByte(path[nodeIdx])
		}
		nodeIdx += 1
	}

	var re, err = trie.getRegex(regex.String())
	if err != nil {
		return false, 0, "", err
	}

	if re != nil && !re.MatchString(val) {
		return true, nodeIdx, name.String(), nil
	} else {
		return false, nodeIdx, name.String(), nil
	}
}

func extractValue(relIdx int, path string) (int, string) {
	var val strings.Builder
	for relIdx < len(path) {
		if path[relIdx] == '/' {
			break
		} else {
			val.WriteByte(path[relIdx])
		}
		relIdx += 1
	}
	return relIdx, val.String()
}

func (trie *Trie[v]) find(method string, path string) (*v, error) {
	nodes := trie.findRoot(method)

	keepSearching := true
	pathIndex := 0
	for keepSearching {
		keepSearching = false
		for i := range *nodes {
			curr := &(*nodes)[i]

			p := 0
			for (pathIndex+p) < len(path) && p < len(curr.path) {
				if path[pathIndex+p] != curr.path[p] {
					break
				}
				p += 1
			}

			if p != 0 {
				if pathIndex+p == len(path) && p == len(curr.path) {
					// case 1: ins path is the same as the curr path - just get the value
					return curr.value, nil
				} else if pathIndex+p == len(path) && p < len(curr.path) {
					// case 2: ins path fits inside the curr path - there is definitely no value here
					return nil, nil
				} else if pathIndex+p < len(path) && p == len(curr.path) {
					// case 3: curr path fits inside the ins path - traverse curr node's children
					nodes = &curr.children
					keepSearching = true
					pathIndex += p
					break
				} else if pathIndex+p < len(path) && p < len(curr.path) {
					// case 4: neither path reaches the end - there is definitely no value here
					return nil, nil
				} else {
					// unkown case
					return nil, errors.New("unknown case for finding node in radix trie: this is a bug")
				}
			}
		}
	}

	return nil, nil
}

func (trie *Trie[v]) routes() []string {
	routes := make([]string, 0)
	for key, root := range trie.roots {
		for _, child := range *root {
			child.routes(key+" "+child.path, &routes)
		}
	}
	return routes
}

func (node *Node[v]) routes(path string, routes *[]string) {
	if node.value != nil {
		*routes = append(*routes, path)
	}
	for i := range node.children {
		child := &node.children[i]
		child.routes(path+child.path, routes)
	}
}
