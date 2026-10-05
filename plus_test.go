package jsonslice

import (
	"bytes"
	"fmt"
	"testing"
)

// TestGetLeavesInputUnchanged checks aggregation does not write into the caller's buffer.
func TestGetLeavesInputUnchanged(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input string
		path  string
		want  string
	}{
		"deep index then terminal slice": {
			input: `[[10,20,30,40],[50,60,70,80]]`,
			path:  `$..[0][1:3]`,
			want:  `[20,30]`,
		},
		"deep index appends after terminal slice": {
			input: `[[[10,20,30,40],0],[7,8,9]]`,
			path:  `$..[0][1:3]`,
			want:  `[0,20,30]`,
		},
		"deep index other bounds": {
			input: `[[10,20,30,40],[50,60,70,80]]`,
			path:  `$..[1][0:2]`,
			want:  `[50,60]`,
		},
		"deep index open end": {
			input: `[[10,20,30,40],[50,60,70,80]]`,
			path:  `$..[0][1:]`,
			want:  `[20,30,40]`,
		},
		"deep negative index then slice": {
			input: `[[10,20,30],[40,50,60],[70,80,90]]`,
			path:  `$..[-1][0:2]`,
			want:  `[70,80]`,
		},
		"deep negative index mid slice": {
			input: `[[10,20,30],[40,50,60],[70,80,90,100]]`,
			path:  `$..[-1][1:3]`,
			want:  `[80,90]`,
		},
		"nested key then terminal slice": {
			input: `[{"nums":[1,2,3,4,5]},{"nums":[6,7,8]}]`,
			path:  `$..[0].nums[1:3]`,
			want:  `[[2,3]]`,
		},
		"deep key then terminal slice": {
			input: `{"book":[{"p":1},{"p":2},{"p":3}]}`,
			path:  `$..book[1:3]`,
			want:  `[[{"p":2},{"p":3}]]`,
		},
		"wildcard field": {
			input: `[{"a":1},{"a":2},{"a":3}]`,
			path:  `$[*].a`,
			want:  `[1,2,3]`,
		},
		"recursive wildcard": {
			input: `{"a":[1,{"b":2}]}`,
			path:  `$..*`,
			want:  `[[1,{"b":2}],1,{"b":2},2]`,
		},
		"terminal slice is returned unchanged": {
			input: `[10,20,30,40]`,
			path:  `$[1:3]`,
			want:  `[20,30]`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assertGetPreservesInput(t, tc.input, tc.path, tc.want)
		})
	}
}

// TestGetSharedInputAcrossPaths evaluates several paths against one buffer.
func TestGetSharedInputAcrossPaths(t *testing.T) {
	t.Parallel()

	raw := `[[[10,20,30,40],0],[7,8,9],{"book":[{"p":1},{"p":2},{"p":3}],"nums":[1,2,3,4]}]`
	paths := []string{
		`$..[0][1:3]`,
		`$..[1][0:2]`,
		`$..[0][1:]`,
		`$..[-1][1:3]`,
		`$..book[1:3]`,
		`$..nums[1:3]`,
		`$[*]`,
		`$..*`,
		`$[1:3]`,
	}

	shared := []byte(raw)
	before := bytes.Clone(shared)
	for _, path := range paths {
		fresh := []byte(raw)
		freshBefore := bytes.Clone(fresh)
		want, wantErr := Get(fresh, path)
		got, gotErr := Get(shared, path)
		if errString(gotErr) != errString(wantErr) {
			t.Fatalf("path %s: shared error %v, fresh error %v", path, gotErr, wantErr)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("path %s: shared %s, fresh %s", path, got, want)
		}
		if !bytes.Equal(fresh, freshBefore) {
			t.Fatalf("path %s wrote into a fresh input", path)
		}
	}
	if !bytes.Equal(shared, before) {
		t.Fatalf("shared input changed\n got %s\nwant %s", shared, before)
	}
}

// errString returns the error text, or an empty string when err is nil.
func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// assertGetPreservesInput checks Get's bytes and that the input buffer is unchanged.
func assertGetPreservesInput(t *testing.T, raw string, path string, want string) {
	t.Helper()

	input := []byte(raw)
	before := bytes.Clone(input)
	got, err := Get(input, path)
	if err != nil {
		t.Fatalf("Get(%s): %v", path, err)
	}
	if !bytes.Equal(input, before) {
		t.Fatalf("input changed\n got %s\nwant %s", input, before)
	}
	if string(got) != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func BenchmarkGetWildcard(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i := 0; i < 200; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(&buf, `{"a":%d}`, i)
	}
	buf.WriteByte(']')
	input := buf.Bytes()

	b.ReportAllocs()
	var sink []byte
	for b.Loop() {
		var err error
		sink, err = Get(input, "$[*].a")
		if err != nil {
			b.Fatal(err)
		}
	}
	if len(sink) == 0 {
		b.Fatal("empty result")
	}
}
