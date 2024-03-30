package httprouter

// a radix trie
type trie[v any] struct {
	roots map[string]*[]node[v]
}

// radix trie nodes
type node[v any] struct {
	path     string
	value    v
	unset    bool
	children []node[v]
}

func newTrie[v any]() trie[v] {
	return trie[v]{
		roots: make(map[string]*[]node[v]),
	}
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
					curr.value = value
					return
				} else if pathIndex+p == len(path) && p < len(curr.path) {
					// case 2: ins path fits inside the curr path - split at where ins path ends
					curr.split(p)
					curr.value = value
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
					curr.unset = true
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
					if curr.unset {
						return nil
					} else {
						return &curr.value
					}
				} else if pathIndex+p == len(path) && p < len(curr.path) {
					// case 2: ins path fits inside the curr path - there is definitely no value here
					return nil
				} else if pathIndex+p < len(path) && p == len(curr.path) {
					// case 3: curr path fits inside the ins path - traverse curr node's children
					nodes = &curr.children
					keepSearching = true
					pathIndex += p
					break
				} else if pathIndex+p < len(path) && p < len(curr.path) {
					// case 4: neither path reaches the end - there is definitely no value here
					return nil
				} else {
					// unkown case
					panic("Unknown case for finding node in radix trie: this is a bug")
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
	if !n.unset {
		*routes = append(*routes, path)
	}
	for i := range n.children {
		child := &n.children[i]
		child.routes(path+child.path, routes)
	}
}
