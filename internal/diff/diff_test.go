// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import (
	"strings"
	"testing"
)

func TestApply(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		edits []Edit
		want  string
	}{
		{"no edits", "hello", nil, "hello"},
		{"single replacement", "hello world", []Edit{{6, 11, "Go"}}, "hello Go"},
		{"insertion", "hello", []Edit{{5, 5, " world"}}, "hello world"},
		{"deletion", "hello world", []Edit{{5, 11, ""}}, "hello"},
		{"multiple edits", "abcdef", []Edit{{1, 2, "B"}, {4, 5, "E"}}, "aBcdEf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Apply(tt.src, tt.edits)
			if err != nil {
				t.Fatalf("Apply() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Apply() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyErrors(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		edits []Edit
		want  string
	}{
		{"out of bounds", "hello", []Edit{{0, 10, "x"}}, "out-of-bounds"},
		{"overlapping", "hello", []Edit{{0, 3, "x"}, {2, 4, "y"}}, "overlapping"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Apply(tt.src, tt.edits)
			if err == nil {
				t.Fatal("Apply() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Apply() error = %q, want to contain %q", err, tt.want)
			}
		})
	}
}

func TestApplyBytes(t *testing.T) {
	src := []byte("hello world")
	edits := []Edit{{6, 11, "Go"}}
	got, err := ApplyBytes(src, edits)
	if err != nil {
		t.Fatalf("ApplyBytes() error = %v", err)
	}
	if string(got) != "hello Go" {
		t.Errorf("ApplyBytes() = %q, want %q", got, "hello Go")
	}
}

func TestSortEdits(t *testing.T) {
	edits := []Edit{
		{10, 15, "c"},
		{0, 5, "a"},
		{5, 10, "b"},
	}
	SortEdits(edits)
	for i := 1; i < len(edits); i++ {
		if edits[i].Start < edits[i-1].Start {
			t.Errorf("edits not sorted: %v", edits)
		}
	}
}

func TestStringsIdentical(t *testing.T) {
	edits := Strings("hello", "hello")
	if edits != nil {
		t.Errorf("Strings() for identical strings = %v, want nil", edits)
	}
}

func TestStringsASCII(t *testing.T) {
	before := "hello world"
	after := "hello Go"
	edits := Strings(before, after)
	if len(edits) == 0 {
		t.Fatal("Strings() returned no edits for different strings")
	}
	got, err := Apply(before, edits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got != after {
		t.Errorf("Apply(Strings()) = %q, want %q", got, after)
	}
}

func TestStringsUnicode(t *testing.T) {
	before := "héllo wörld"
	after := "héllo Gö"
	edits := Strings(before, after)
	if len(edits) == 0 {
		t.Fatal("Strings() returned no edits for different strings")
	}
	got, err := Apply(before, edits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got != after {
		t.Errorf("Apply(Strings()) = %q, want %q", got, after)
	}
}

func TestBytesIdentical(t *testing.T) {
	edits := Bytes([]byte("hello"), []byte("hello"))
	if edits != nil {
		t.Errorf("Bytes() for identical input = %v, want nil", edits)
	}
}

func TestBytesASCII(t *testing.T) {
	before := []byte("abc")
	after := []byte("aXc")
	edits := Bytes(before, after)
	got, err := ApplyBytes(before, edits)
	if err != nil {
		t.Fatalf("ApplyBytes() error = %v", err)
	}
	if string(got) != "aXc" {
		t.Errorf("ApplyBytes(Bytes()) = %q, want %q", got, "aXc")
	}
}

func TestUnifiedNoDiff(t *testing.T) {
	result := Unified("a.txt", "b.txt", "same\n", "same\n")
	if result != "" {
		t.Errorf("Unified() for identical strings = %q, want empty", result)
	}
}

func TestUnifiedSimple(t *testing.T) {
	old := "line1\nline2\nline3\n"
	new := "line1\nmodified\nline3\n"
	result := Unified("old.txt", "new.txt", old, new)
	if !strings.Contains(result, "--- old.txt") {
		t.Error("missing old label")
	}
	if !strings.Contains(result, "+++ new.txt") {
		t.Error("missing new label")
	}
	if !strings.Contains(result, "-line2") {
		t.Error("missing deletion")
	}
	if !strings.Contains(result, "+modified") {
		t.Error("missing insertion")
	}
}

func TestToUnifiedZeroContext(t *testing.T) {
	old := "line1\nline2\nline3\nline4\nline5\n"
	new := "line1\nline2\nchanged\nline4\nline5\n"
	edits := Strings(old, new)
	result, err := ToUnified("old", "new", old, edits, 0)
	if err != nil {
		t.Fatalf("ToUnified() error = %v", err)
	}
	// With 0 context lines, unchanged lines should not appear
	if strings.Contains(result, " line1") {
		t.Error("context line1 should not appear with 0 context")
	}
	if !strings.Contains(result, "-line3") {
		t.Error("missing deletion of line3")
	}
	if !strings.Contains(result, "+changed") {
		t.Error("missing insertion of changed")
	}
}

func TestToUnifiedNoEdits(t *testing.T) {
	result, err := ToUnified("old", "new", "content\n", nil, 3)
	if err != nil {
		t.Fatalf("ToUnified() error = %v", err)
	}
	if result != "" {
		t.Errorf("ToUnified() with no edits = %q, want empty", result)
	}
}

func TestMergeIdentical(t *testing.T) {
	x := []Edit{{0, 5, "hello"}}
	y := []Edit{{0, 5, "hello"}}
	merged, ok := Merge(x, y)
	if !ok {
		t.Fatal("Merge() returned conflict for identical edits")
	}
	if len(merged) != 1 {
		t.Fatalf("Merge() returned %d edits, want 1", len(merged))
	}
	if merged[0] != x[0] {
		t.Errorf("Merge() = %v, want %v", merged[0], x[0])
	}
}

func TestMergeNonOverlapping(t *testing.T) {
	x := []Edit{{0, 3, "aaa"}}
	y := []Edit{{5, 8, "bbb"}}
	merged, ok := Merge(x, y)
	if !ok {
		t.Fatal("Merge() returned conflict for non-overlapping edits")
	}
	if len(merged) != 2 {
		t.Fatalf("Merge() returned %d edits, want 2", len(merged))
	}
}

func TestMergeConflict(t *testing.T) {
	x := []Edit{{0, 5, "hello"}}
	y := []Edit{{0, 5, "world"}}
	_, ok := Merge(x, y)
	if ok {
		t.Error("Merge() should return conflict for different edits at same position")
	}
}

func TestMergeEmpty(t *testing.T) {
	x := []Edit{{0, 3, "aaa"}}
	merged, ok := Merge(x, nil)
	if !ok {
		t.Fatal("Merge() returned conflict")
	}
	if len(merged) != 1 {
		t.Fatalf("Merge() returned %d edits, want 1", len(merged))
	}
}

func TestLinesIdentical(t *testing.T) {
	edits := Lines("hello\n", "hello\n")
	if edits != nil {
		t.Errorf("Lines() for identical strings = %v, want nil", edits)
	}
}

func TestLinesSimple(t *testing.T) {
	before := "line1\nline2\nline3\n"
	after := "line1\nmodified\nline3\n"
	edits := Lines(before, after)
	if len(edits) == 0 {
		t.Fatal("Lines() returned no edits for different strings")
	}
	got, err := Apply(before, edits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got != after {
		t.Errorf("Apply(Lines()) = %q, want %q", got, after)
	}
}

func TestLinesInsert(t *testing.T) {
	before := "line1\nline3\n"
	after := "line1\nline2\nline3\n"
	edits := Lines(before, after)
	got, err := Apply(before, edits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got != after {
		t.Errorf("Apply(Lines()) = %q, want %q", got, after)
	}
}

func TestLinesDelete(t *testing.T) {
	before := "line1\nline2\nline3\n"
	after := "line1\nline3\n"
	edits := Lines(before, after)
	got, err := Apply(before, edits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got != after {
		t.Errorf("Apply(Lines()) = %q, want %q", got, after)
	}
}

func TestLinesToUnified(t *testing.T) {
	before := "line1\nline2\nline3\n"
	after := "line1\nchanged\nline3\n"
	edits := Lines(before, after)
	result, err := ToUnified("before.txt", "after.txt", before, edits, 0)
	if err != nil {
		t.Fatalf("ToUnified() error = %v", err)
	}
	if !strings.Contains(result, "-line2") {
		t.Error("missing deletion of line2")
	}
	if !strings.Contains(result, "+changed") {
		t.Error("missing insertion of changed")
	}
}

// TestRoundTrip verifies that applying edits from Strings produces the expected result.
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		before string
		after  string
	}{
		{"empty to nonempty", "", "hello\n"},
		{"nonempty to empty", "hello\n", ""},
		{"single line change", "aaa\n", "bbb\n"},
		{"multiline", "a\nb\nc\n", "a\nB\nc\n"},
		{"add lines", "a\n", "a\nb\nc\n"},
		{"remove lines", "a\nb\nc\n", "a\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edits := Strings(tt.before, tt.after)
			got, err := Apply(tt.before, edits)
			if err != nil {
				t.Fatalf("Apply() error = %v", err)
			}
			if got != tt.after {
				t.Errorf("round-trip failed: got %q, want %q", got, tt.after)
			}
		})
	}
}
