package httprouter

type trie[v any] struct {
	roots map[string]*[]node[v]
}

func newTrie[v any]() trie[v] {
	return trie[v]{
		roots: make(map[string]*[]node[v]),
	}
}

type node[v any] struct {
	path     string
	value    v
	children []node[v]
}

func (trie *trie[v]) findRoot(method string) *[]node[v] {
	nodes, ok := trie.roots[method]
	if !ok {
		nodes := make([]node[v], 0)
		trie.roots[method] = &nodes
		return &nodes
	}
	return nodes
}

func (trie *trie[v]) insert(method string, path string, value v) {
	nodes := trie.findRoot(method)

	keepSearching := true
	pathIndex := 0
	for keepSearching {
		keepSearching = false
		for i := range *nodes {
			curr := &(*nodes)[i]
			pathsMatch := true
			ogPathIndex := pathIndex
			nodePathIndex := 0

			for pathIndex < len(path) && nodePathIndex < len(curr.path) {
				if path[pathIndex] != curr.path[nodePathIndex] {
					pathsMatch = false
					break
				}
				pathIndex += 1
				nodePathIndex += 1
			}

			if !pathsMatch {
				pathIndex = ogPathIndex
			} else {
				if pathIndex == len(path) && nodePathIndex == len(curr.path) {
					curr.value = value
					return
				} else if nodePathIndex < len(curr.path) {
					curr.split(nodePathIndex)
					curr.value = value
					return
				} else {
					nodes = &curr.children
					keepSearching = true
					break
				}
			}
		}
	}

	*nodes = append(*nodes, newNode[v](path[pathIndex:], value))
}

func newNode[v any](path string, value v) node[v] {
	return node[v]{
		path:     path,
		children: make([]node[v], 0),
		value:    value,
	}
}

func (n *node[v]) split(splitIndex int) {
	next := node[v]{
		path:     n.path[splitIndex:],
		value:    n.value,
		children: n.children,
	}
	n.path = n.path[:splitIndex]
	n.children = []node[v]{next}
}

func (trie *trie[v]) find(method string, path string) *v {
	nodes := trie.findRoot(method)

	keepSearching := true
	pathIndex := 0
	for keepSearching {
		keepSearching = false
		for i := range *nodes {
			curr := &(*nodes)[i]
			pathsMatch := true
			ogPathIndex := pathIndex
			nodePathIndex := 0

			for pathIndex < len(path) && nodePathIndex < len(curr.path) {
				if path[pathIndex] != curr.path[nodePathIndex] {
					pathsMatch = false
					break
				}
				pathIndex += 1
				nodePathIndex += 1
			}

			if !pathsMatch {
				pathIndex = ogPathIndex
			} else {
				if pathIndex == len(path) && nodePathIndex == len(curr.path) {
					return &curr.value
				} else if nodePathIndex < len(curr.path) {
					return nil
				} else {
					nodes = &curr.children
					keepSearching = true
					break
				}
			}
		}
	}

	return nil
}

func (trie *trie[v]) routes() []string {
	routes := make([]string, 0)
	for key, root := range trie.roots {
		for _, child := range *root {
			child.routes(key+" "+child.path, &routes)
		}
	}
	return routes
}

func (n *node[v]) routes(path string, routes *[]string) {
	*routes = append(*routes, path)
	for i := range n.children {
		child := &n.children[i]
		child.routes(path+child.path, routes)
	}
}
