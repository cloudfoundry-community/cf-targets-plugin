// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lcs

import (
	"testing"
)

func TestDiffStringsIdentical(t *testing.T) {
	diffs := DiffStrings("hello", "hello")
	if len(diffs) != 0 {
		t.Errorf("DiffStrings() for identical = %v, want empty", diffs)
	}
}

func TestDiffStringsSimple(t *testing.T) {
	diffs := DiffStrings("abc", "aXc")
	if len(diffs) == 0 {
		t.Fatal("DiffStrings() returned no diffs")
	}
	// The diff should replace 'b' (offset 1-2) with 'X' (offset 1-2)
	d := diffs[0]
	if d.Start != 1 || d.End != 2 {
		t.Errorf("DiffStrings() deletion range = [%d,%d), want [1,2)", d.Start, d.End)
	}
	if d.ReplStart != 1 || d.ReplEnd != 2 {
		t.Errorf("DiffStrings() replacement range = [%d,%d), want [1,2)", d.ReplStart, d.ReplEnd)
	}
}

func TestDiffStringsCompleteDifference(t *testing.T) {
	diffs := DiffStrings("abc", "xyz")
	if len(diffs) == 0 {
		t.Fatal("DiffStrings() returned no diffs for completely different strings")
	}
}

func TestDiffBytesIdentical(t *testing.T) {
	diffs := DiffBytes([]byte("hello"), []byte("hello"))
	if len(diffs) != 0 {
		t.Errorf("DiffBytes() for identical = %v, want empty", diffs)
	}
}

func TestDiffBytesSimple(t *testing.T) {
	diffs := DiffBytes([]byte("abcd"), []byte("aXYd"))
	if len(diffs) == 0 {
		t.Fatal("DiffBytes() returned no diffs")
	}
	d := diffs[0]
	if d.Start != 1 || d.End != 3 {
		t.Errorf("DiffBytes() deletion range = [%d,%d), want [1,3)", d.Start, d.End)
	}
}

func TestDiffRunesIdentical(t *testing.T) {
	diffs := DiffRunes([]rune("hello"), []rune("hello"))
	if len(diffs) != 0 {
		t.Errorf("DiffRunes() for identical = %v, want empty", diffs)
	}
}

func TestDiffRunesSimple(t *testing.T) {
	diffs := DiffRunes([]rune("héllo"), []rune("hëllo"))
	if len(diffs) == 0 {
		t.Fatal("DiffRunes() returned no diffs for different rune strings")
	}
}

func TestDiffLinesIdentical(t *testing.T) {
	diffs := DiffLines([]string{"a\n", "b\n"}, []string{"a\n", "b\n"})
	if len(diffs) != 0 {
		t.Errorf("DiffLines() for identical = %v, want empty", diffs)
	}
}

func TestDiffLinesSimple(t *testing.T) {
	a := []string{"line1\n", "line2\n", "line3\n"}
	b := []string{"line1\n", "changed\n", "line3\n"}
	diffs := DiffLines(a, b)
	if len(diffs) == 0 {
		t.Fatal("DiffLines() returned no diffs")
	}
	d := diffs[0]
	if d.Start != 1 || d.End != 2 {
		t.Errorf("DiffLines() deletion range = [%d,%d), want [1,2)", d.Start, d.End)
	}
	if d.ReplStart != 1 || d.ReplEnd != 2 {
		t.Errorf("DiffLines() replacement range = [%d,%d), want [1,2)", d.ReplStart, d.ReplEnd)
	}
}

func TestDiffLinesInsert(t *testing.T) {
	a := []string{"line1\n", "line3\n"}
	b := []string{"line1\n", "line2\n", "line3\n"}
	diffs := DiffLines(a, b)
	if len(diffs) == 0 {
		t.Fatal("DiffLines() returned no diffs for insertion")
	}
}

func TestDiffLinesDelete(t *testing.T) {
	a := []string{"line1\n", "line2\n", "line3\n"}
	b := []string{"line1\n", "line3\n"}
	diffs := DiffLines(a, b)
	if len(diffs) == 0 {
		t.Fatal("DiffLines() returned no diffs for deletion")
	}
}

func TestDiffStringsEmpty(t *testing.T) {
	tests := []struct {
		name string
		a, b string
	}{
		{"both empty", "", ""},
		{"a empty", "", "hello"},
		{"b empty", "hello", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs := DiffStrings(tt.a, tt.b)
			// Apply the diffs to verify correctness
			result := tt.a
			// Reconstruct from diffs
			var out []byte
			pos := 0
			for _, d := range diffs {
				out = append(out, tt.a[pos:d.Start]...)
				out = append(out, tt.b[d.ReplStart:d.ReplEnd]...)
				pos = d.End
			}
			out = append(out, tt.a[pos:]...)
			result = string(out)
			if result != tt.b {
				t.Errorf("applying diffs: got %q, want %q", result, tt.b)
			}
		})
	}
}
