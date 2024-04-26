package httprouter

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestTrieInsertPaths(t *testing.T) {
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
	trie.insert("GET", "/hell/today", 10)
	trie.insert("GET", "/food", 11)

	expectedNodes := []Node[int]{
		{path: "/",
			children: []Node[int]{
				{path: "foo", value: intPtr(0),
					children: []Node[int]{
						{path: "/bar", value: intPtr(8), children: []Node[int]{}},
						{path: "d", value: intPtr(11), children: []Node[int]{}},
					},
				},
				{path: "he", value: intPtr(5),
					children: []Node[int]{
						{path: "ll", value: intPtr(1),
							children: []Node[int]{
								{path: "o", value: intPtr(3),
									children: []Node[int]{
										{path: "/",
											children: []Node[int]{
												{path: "world", value: intPtr(4), children: []Node[int]{}},
												{path: "name", value: intPtr(6), children: []Node[int]{}},
											},
										},
									},
								},
								{path: "/today", value: intPtr(10), children: []Node[int]{}},
							},
						},
						{path: "y", value: intPtr(7), children: []Node[int]{}},
					},
				},
			},
		},
	}

	branch, ok := trie.roots["GET"]
	if !ok {
		t.Errorf("Expected a GET branch at the root")
	}

	expectedStr := debugNodes(expectedNodes, 0)
	actualStr := debugNodes(*branch, 0)
	if expectedStr != actualStr {
		t.Errorf("Expected %s \n\nfor the Trie structure but got %s", expectedStr, actualStr)
	}
}

func TestTrieInsertPrefixes(t *testing.T) {
	trie := newTrie[int]()

	trie.insert("GET", "test", 0)
	trie.insert("GET", "slow", 1)
	trie.insert("GET", "water", 2)
	trie.insert("GET", "slower", 3)
	trie.insert("GET", "tester", 4)
	trie.insert("GET", "team", 5)
	trie.insert("GET", "toast", 6)

	expectedNodes := []Node[int]{
		{path: "t",
			children: []Node[int]{
				{path: "e",
					children: []Node[int]{
						{path: "st", value: intPtr(0),
							children: []Node[int]{
								{path: "er", value: intPtr(4), children: []Node[int]{}},
							},
						},
						{path: "am", value: intPtr(5), children: []Node[int]{}},
					},
				},
				{path: "oast", value: intPtr(6), children: []Node[int]{}},
			},
		},
		{path: "slow", value: intPtr(1),
			children: []Node[int]{
				{path: "er", value: intPtr(3), children: []Node[int]{}},
			},
		},
		{path: "water", value: intPtr(2), children: []Node[int]{}},
	}

	branch, ok := trie.roots["GET"]
	if !ok {
		t.Errorf("Expected a GET branch at the root")
	}

	expectedStr := debugNodes(expectedNodes, 0)
	actualStr := debugNodes(*branch, 0)
	if expectedStr != actualStr {
		t.Errorf("Expected %s \n\nfor the Trie structure but got %s", expectedStr, actualStr)
	}
}

func intPtr(i int) *int {
	return &i
}

func TestTrieFind(t *testing.T) {
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
		{in: "/", out: nil},
		{in: "/foo", out: intPtr(0)},
		{in: "/hello/world", out: intPtr(4)},
		{in: "/foo/bar", out: intPtr(8)},
		{in: "/heyo", out: nil},
		{in: "/hey", out: intPtr(7)},
		{in: "/foo/bar1", out: nil},
	}

	for _, test := range testTable {
		value, err := trie.find("GET", test.in)
		if err != nil {
			t.Errorf("Error for path %s, got %v", test.in, err)
		}

		if value != nil || test.out != nil {
			if value == nil {
				t.Errorf("Expected to find %v in the Trie structure for path %s but got nil", *test.out, test.in)
			} else if *value != *test.out {
				t.Errorf("Expected to find %v in the Trie structure for path %s but got %v", *test.out, test.in, *value)
			}
		}
	}
}

func TestTrieRoutes(t *testing.T) {
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
		t.Errorf("Expected %v for the Trie structure but got a %v", expectedRoutes, routes)
	}
}

func randStrings(count int) []string {
	minLen := 5
	maxLen := 15
	chars := "abcdefghijklmnopqrstuvwxyz"
	strLen := rand.Intn(maxLen) + minLen

	strs := make([]string, 0)
	for i := 0; i < count; i++ {
		var sb strings.Builder
		for k := 0; k < strLen; k++ {
			c := rand.Intn(len(chars))
			sb.WriteByte(chars[c])
		}
		strs = append(strs, sb.String())
	}
	return strs
}

func generatePaths() []string {
	routeCount := 10000
	routesLen := 50

	builders := make([]strings.Builder, routeCount)
	for i := 0; i < routesLen; i++ {
		strsCount := i/5 + 1
		strs := randStrings(strsCount)

		for j := range builders {
			rb := &builders[j]
			rb.WriteByte('/')
			r := rand.Intn(len(strs))
			rb.WriteString(strs[r])
		}
	}

	routes := make([]string, 0)
	for i := range builders {
		routes = append(routes, builders[i].String())
	}

	return routes
}

func BenchmarkTrie(b *testing.B) {
	paths := generatePaths()
	trie := newTrie[int]()

	b.StartTimer()

	for i, route := range paths {
		trie.insert("GET", route, i)
	}

	b.StopTimer()
	b.Logf("Trie insert took %d ms", b.Elapsed()/time.Millisecond)

	b.ResetTimer()
	b.StartTimer()

	for _, path := range paths {
		_, err := trie.find("GET", path)
		if err != nil {
			b.Fatalf("Error occured: %v", err)
		}
	}

	b.StopTimer()
	b.Logf("Trie find took %d ms", b.Elapsed()/time.Millisecond)
}
