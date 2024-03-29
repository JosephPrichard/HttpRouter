package httprouter

import (
	"reflect"
	"testing"
)

func TestInsertTrie(t *testing.T) {
	trie := newTrie[int]()

	trie.insert("GET", "/foo", 0)
	trie.insert("GET", "/hell", 1)
	trie.insert("GET", "/hello/world", 2)
	trie.insert("GET", "/hello", 3)
	trie.insert("GET", "/hello/world", 4)
	trie.insert("GET", "/he", 5)
	trie.insert("GET", "/hello/name", 6)
	trie.insert("GET", "/hey", 7)
	trie.insert("GET", "/foo/bar", 8)

	expectedNodes := []node[int]{
		{path: "/foo", value: 0, children: []node[int]{{path: "/bar", value: 8, children: []node[int]{}}}},
		{path: "/he", value: 5, children: []node[int]{
			{path: "ll", value: 1, children: []node[int]{{
				path: "o", value: 3, children: []node[int]{
					{path: "/world", value: 4, children: []node[int]{}},
					{path: "/name", value: 6, children: []node[int]{}},
				},
			}}},
			{path: "y", value: 7, children: []node[int]{}},
		},
		},
	}

	branch, ok := trie.roots["GET"]
	if !ok {
		t.Errorf("Expected a GET branch at the root")
	}

	if !reflect.DeepEqual(expectedNodes, *branch) {
		t.Errorf("Expected %v for the trie structure but got a %v", expectedNodes, *branch)
	}
}

func intPtr(i int) *int {
	return &i
}

func TestFindTrie(t *testing.T) {
	trie := newTrie[int]()

	trie.insert("GET", "/foo", 0)
	trie.insert("GET", "/hell", 1)
	trie.insert("GET", "/hello/world", 2)
	trie.insert("GET", "/hello", 3)
	trie.insert("GET", "/hello/world", 4)
	trie.insert("GET", "/he", 5)
	trie.insert("GET", "/hello/name", 6)
	trie.insert("GET", "/hey", 7)
	trie.insert("GET", "/foo/bar", 8)

	type Test struct {
		in  string
		out *int
	}

	testTable := []Test{
		{in: "/foo", out: intPtr(0)},
		{in: "/hello/world", out: intPtr(4)},
		{in: "/foo/bar", out: intPtr(8)},
		{in: "/foo/bar1", out: nil},
	}

	for _, test := range testTable {
		value := trie.find("GET", test.in)
		if value == nil && test.out == nil {
			continue
		}
		if *value == *test.out {
			continue
		}
		t.Errorf("Expected to find %v in the trie structure for path %s but got %v", *test.out, test.in, *value)
	}
}

func TestRoutesTrie(t *testing.T) {
	trie := newTrie[int]()

	trie.insert("GET", "/foo", 0)
	trie.insert("GET", "/hell", 1)
	trie.insert("GET", "/hello/world", 2)
	trie.insert("GET", "/hello", 3)
	trie.insert("GET", "/hello/world", 4)
	trie.insert("GET", "/he", 5)
	trie.insert("GET", "/hello/name", 6)
	trie.insert("GET", "/hey", 7)
	trie.insert("GET", "/foo/bar", 8)

	routes := trie.routes()

	expectedRoutes := []string{
		"GET /foo",
		"GET /foo/bar",
		"GET /he",
		"GET /hell",
		"GET /hello",
		"GET /hello/world",
		"GET /hello/name",
		"GET /hey",
	}

	if !reflect.DeepEqual(expectedRoutes, routes) {
		t.Errorf("Expected %v for the trie structure but got a %v", expectedRoutes, routes)
	}
}
