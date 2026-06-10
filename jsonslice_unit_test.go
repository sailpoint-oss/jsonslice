package jsonslice

import (
	"testing"
)

// TestGetEmbeddedPaths covers $-root paths referenced inside [?( )] filters only (not {{ }} templates
// or $.x used as an array index). Paths use generic names (foo, bar, a, b, …) but match real shapes.
func TestGetEmbeddedPaths(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantNil   bool
		wantPaths []string // order ignored; empty means expect non-nil slice with len 0 when wantNil is false
	}{
		{
			name:    "empty path",
			path:    "",
			wantNil: true,
		},
		{
			name:    "only root",
			path:    "$",
			wantNil: true,
		},
		{
			name:    "does not start with dollar",
			path:    "foo",
			wantNil: true,
		},
		{
			name:    "malformed root subscript",
			path:    "$.[",
			wantNil: true,
		},
		{
			name:    "go template index brackets not filter embeds",
			path:    `$.foo[{{$.bar}}]`,
			wantNil: true,
		},
		{
			name:    "dollar in array index bracket not filter variable",
			path:    `$.foo.bar[$.baz.qux].name`,
			wantNil: true,
		},
		{
			name:      "property path no filter",
			path:      "$.foo.bar",
			wantPaths: []string{},
		},
		{
			name:      "filter with only at references",
			path:      `$.foo[?(@.a==1)]`,
			wantPaths: []string{},
		},
		{
			name:      "filter embeds dollar path after property compare",
			path:      `$.foo.bar.baz[?(@.k==$.a.b.c)].d`,
			wantPaths: []string{"$.a.b.c"},
		},
		{
			name:      "same embed one fewer outer segment",
			path:      `$.foo.bar[?(@.k==$.a.b.c)].d`,
			wantPaths: []string{"$.a.b.c"},
		},
		{
			name:      "spaces around equality before embedded path",
			path:      `$.foo.bar[?(@.x == $.a.b)].c`,
			wantPaths: []string{"$.a.b"},
		},
		{
			name:      "spaces around equality different left property",
			path:      `$.foo.bar[?(@.y == $.c.d)]`,
			wantPaths: []string{"$.c.d"},
		},
		{
			name:      "filter embeds path used as inner dependency",
			path:      `$.foo.items[?(@.id==$.bar.baz)].name`,
			wantPaths: []string{"$.bar.baz"},
		},
		{
			name:      "two filters in sequence each embed dollar path",
			path:      `$.a[?(@.x==$.b)].c[?(@.y==$.d)]`,
			wantPaths: []string{"$.b", "$.d"},
		},
		{
			name:      "filter embeds indexed dollar path",
			path:      `$.foo.items[?(@.code != $.bar[0].code)].value`,
			wantPaths: []string{"$.bar[0].code"},
		},
		{
			name:      "same indexed embed longer outer path",
			path:      `$.baz.qux.items[?(@.code != $.bar[0].code)].value`,
			wantPaths: []string{"$.bar[0].code"},
		},
		{
			name:      "two distinct dollar paths in one filter",
			path:      `$.foo[?(@.a==$.b && @.c==$.d)]`,
			wantPaths: []string{"$.b", "$.d"},
		},
		{
			name:      "repeated dollar path deduplicated",
			path:      `$.foo[?(@.a==$.b.c && @.a==$.b.c)]`,
			wantPaths: []string{"$.b.c"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetEmbeddedPaths(tc.path)
			if tc.wantNil {
				if got != nil {
					t.Fatalf("GetEmbeddedPaths(%q) = %v, want nil", tc.path, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("GetEmbeddedPaths(%q) = nil, want non-nil slice", tc.path)
			}
			if len(tc.wantPaths) == 0 {
				if len(got) != 0 {
					t.Fatalf("GetEmbeddedPaths(%q) = %v, want empty slice", tc.path, got)
				}
				return
			}
			if len(got) != len(tc.wantPaths) {
				t.Fatalf("GetEmbeddedPaths(%q) = %v (len %d), want len %d %v", tc.path, got, len(got), len(tc.wantPaths), tc.wantPaths)
			}
			used := make([]bool, len(got))
			for _, w := range tc.wantPaths {
				found := false
				for i := range got {
					if !used[i] && got[i] == w {
						used[i] = true
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("GetEmbeddedPaths(%q) = %v, missing or extra mismatch for want %q (want %v)", tc.path, got, w, tc.wantPaths)
				}
			}
		})
	}
}

// TestGetFilterRelativePaths covers @-relative paths referenced inside [?( )] filters only.
// It mirrors TestGetEmbeddedPaths but asserts on the current-element ("@") operands instead of
// the root ("$") operands. Order is significant: paths are returned in first-seen order with
// duplicates removed.
func TestGetFilterRelativePaths(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantNil   bool
		wantPaths []string // order significant; empty means expect non-nil slice with len 0 when wantNil is false
	}{
		{
			name:    "empty path",
			path:    "",
			wantNil: true,
		},
		{
			name:    "only root",
			path:    "$",
			wantNil: true,
		},
		{
			name:    "does not start with dollar",
			path:    "foo",
			wantNil: true,
		},
		{
			name:    "malformed root subscript",
			path:    "$.[",
			wantNil: true,
		},
		{
			name:    "go template index brackets are a parse error",
			path:    `$.foo[{{$.bar}}]`,
			wantNil: true,
		},
		{
			name:      "property path no filter",
			path:      "$.foo.bar",
			wantPaths: []string{},
		},
		{
			name:      "filter with single at reference",
			path:      `$.foo[?(@.a==1)]`,
			wantPaths: []string{"@.a"},
		},
		{
			name:      "filter compares at reference to dollar path",
			path:      `$.foo.bar.baz[?(@.k==$.a.b.c)].d`,
			wantPaths: []string{"@.k"},
		},
		{
			name:      "two at references in one filter preserve order",
			path:      `$.foo[?(@.a==$.b && @.c==$.d)]`,
			wantPaths: []string{"@.a", "@.c"},
		},
		{
			name:      "two at references compared to literals",
			path:      `$.foo[?(@.a==1 && @.b==2)]`,
			wantPaths: []string{"@.a", "@.b"},
		},
		{
			name:      "two filters in sequence each contribute one at reference",
			path:      `$.a[?(@.x==1)].c[?(@.y==2)]`,
			wantPaths: []string{"@.x", "@.y"},
		},
		{
			name:      "repeated at reference deduplicated",
			path:      `$.foo[?(@.a==1 && @.a==2)]`,
			wantPaths: []string{"@.a"},
		},
		{
			name:      "quoted at reference key",
			path:      `$.items[?(@['first name']=="x")]`,
			wantPaths: []string{"@['first name']"},
		},
		{
			name:      "indexed at reference",
			path:      `$.foo[?(@.a[0].b==1)]`,
			wantPaths: []string{"@.a[0].b"},
		},
		{
			name:      "bare at reference",
			path:      `$.foo[?(@==1)]`,
			wantPaths: []string{"@"},
		},
		{
			name:      "filter with only dollar reference has no relatives",
			path:      `$.foo[?($.a.b==1)]`,
			wantPaths: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetFilterRelativePaths(tc.path)
			if tc.wantNil {
				if got != nil {
					t.Fatalf("GetFilterRelativePaths(%q) = %v, want nil", tc.path, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("GetFilterRelativePaths(%q) = nil, want non-nil slice", tc.path)
			}
			if len(got) != len(tc.wantPaths) {
				t.Fatalf("GetFilterRelativePaths(%q) = %v (len %d), want %v (len %d)", tc.path, got, len(got), tc.wantPaths, len(tc.wantPaths))
			}
			for i := range tc.wantPaths {
				if got[i] != tc.wantPaths[i] {
					t.Fatalf("GetFilterRelativePaths(%q) = %v, want %v (mismatch at index %d)", tc.path, got, tc.wantPaths, i)
				}
			}
		})
	}
}

func Test_AdjustBounds(t *testing.T) {
	type Input struct{ left, right, step int }
	type Expected struct {
		a, b, step int
		err        bool
	}
	
	tests := []struct {
		Input
		Expected
	}{
		// slice is [0,1,2,3,4]
		{Input{cEmpty, cEmpty, cEmpty}, Expected{0, 5, 1, false}}, // [:] -> [0..5)
		{Input{2, cEmpty, cEmpty}, Expected{2, 5, 1, false}},      // [2:] -> [2..5)
		{Input{cEmpty, 3, cEmpty}, Expected{0, 3, 1, false}},      // [:3] -> [0..3)
		{Input{-2, cEmpty, cEmpty}, Expected{3, 5, 1, false}},     // [-2:] -> [3..5)
		{Input{cEmpty, -2, cEmpty}, Expected{0, 3, 1, false}},     // [:-2] -> [0..3)
		{Input{1, 4, cEmpty}, Expected{1, 4, 1, false}},           // [1:4] -> [1..4)
		{Input{-4, 4, cEmpty}, Expected{1, 4, 1, false}},          // [-4:4] -> [1..4)
		{Input{-8, 4, cEmpty}, Expected{0, 4, 1, false}},          // [-8:4] -> [0..4)
		{Input{-8, 8, cEmpty}, Expected{0, 5, 1, false}},          // [-8:8] -> [0..5)
		{Input{1, -2, cEmpty}, Expected{1, 3, 1, false}},          // [1:-2] -> [1..3)
		{Input{1, -3, cEmpty}, Expected{1, 2, 1, false}},          // [1:-3] -> [1..2)
		{Input{1, -4, cEmpty}, Expected{1, 1, 1, false}},          // [1:-4] -> [1..1)
		{Input{-5, -2, cEmpty}, Expected{0, 3, 1, false}},         // [-5:-2] -> [0..3)
	}

	n := 5 // slice length
	
	for _, tst := range tests {
		a, b, step, err := adjustBounds(tst.Input.left, tst.Input.right, tst.Input.step, n)
		if !(a == tst.Expected.a && b == tst.Expected.b && step == tst.Expected.step) || (err != nil) != tst.Expected.err {
			t.Errorf(
				"adjustBounds(%v) == {%v,%v,%v,%v}, expected %v",
				tst.Input, a, b, step, err, tst.Expected,
			)
		}
	}
}

func Test_sliceRecurse(t *testing.T) {
	input := []byte(`["a","b","c","d","e"]`)
	elems := []tElem{{1, 4}, {5, 8}, {9, 12}, {13, 16}, {17, 20}}
	tests := []struct {
		nod      *tNode
		expected string
	}{
		
		{&tNode{Slice: [3]int{cEmpty, cEmpty, cEmpty}}, `"a","b","c","d","e"`}, // [:] == [::]
		{&tNode{Slice: [3]int{2, cEmpty, cEmpty}}, `"c","d","e"`},              // [2:]
		{&tNode{Slice: [3]int{cEmpty, 3, cEmpty}}, `"a","b","c"`},              // [:3]
		{&tNode{Slice: [3]int{-2, cEmpty, cEmpty}}, `"d","e"`},                 // [-2:]
		{&tNode{Slice: [3]int{cEmpty, -2, cEmpty}}, `"a","b","c"`},             // [:-2]
		{&tNode{Slice: [3]int{1, 4, cEmpty}}, `"b","c","d"`},                   // [1:4]
		{&tNode{Slice: [3]int{-4, 4, cEmpty}}, `"b","c","d"`},                  // [-4:4]
		{&tNode{Slice: [3]int{-8, 4, cEmpty}}, `"a","b","c","d"`},              // [-8:4]
		{&tNode{Slice: [3]int{-8, 8, cEmpty}}, `"a","b","c","d","e"`},          // [-8:8]
		{&tNode{Slice: [3]int{1, -2, cEmpty}}, `"b","c"`},                      // [1:-2]
		{&tNode{Slice: [3]int{1, -3, cEmpty}}, `"b"`},                          // [1:-3]
		{&tNode{Slice: [3]int{1, -4, cEmpty}}, ``},                             // [1:-4]
		{&tNode{Slice: [3]int{-5, -2, cEmpty}}, `"a","b","c"`},                 // [-5:-2] -> 0..2

		// slice + step
		{&tNode{Slice: [3]int{cEmpty, cEmpty, 1}}, `"a","b","c","d","e"`}, // [::1]
		{&tNode{Slice: [3]int{0, cEmpty, 2}}, `"a","c","e"`},              // [::2]
		{&tNode{Slice: [3]int{1, cEmpty, 2}}, `"b","d"`},                  // [1::2]

		// slice + negative step
		{&tNode{Slice: [3]int{cEmpty, cEmpty, -1}}, `"e","d","c","b","a"`}, // [::-1]
		{&tNode{Slice: [3]int{2, cEmpty, -1}}, `"c","b","a"`},              // [2::-1]
		{&tNode{Slice: [3]int{cEmpty, 3, -1}}, `"e"`},                      // [:3:-1]
		{&tNode{Slice: [3]int{-2, cEmpty, -1}}, `"d","c","b","a"`},         // [-2::-1]
		{&tNode{Slice: [3]int{-2, 1, -1}}, `"d","c"`},                      // [-2:1:-1]
		{&tNode{Slice: [3]int{cEmpty, cEmpty, -2}}, `"e","c","a"`},         // [::-2]
		{&tNode{Slice: [3]int{cEmpty, cEmpty, -3}}, `"e","b"`},             // [::-3]
		{&tNode{Slice: [3]int{cEmpty, cEmpty, -4}}, `"e","a"`},             // [::-4]
		{&tNode{Slice: [3]int{cEmpty, cEmpty, -5}}, `"e"`},                 // [::-5]
		{&tNode{Slice: [3]int{1, 2, -1}}, ``},                              // [1:2:-1]
	}

	for _, tst := range tests {
		tst.nod.Type |= cFullScan
		res, err := sliceRecurse(input, tst.nod, elems)
		if err != nil || tst.expected != string(res) {
			t.Errorf(
				"sliceRecurse('%v', {%d, %d, %d}) == %v, expected %v",
				string(input), tst.nod.Slice[0], tst.nod.Slice[1], tst.nod.Slice[2], string(res), tst.expected,
			)
		}
	}
}
