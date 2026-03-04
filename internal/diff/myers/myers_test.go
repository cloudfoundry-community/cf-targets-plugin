// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package myers

import (
	"testing"

	"github.com/norman-abramovitz/cf-targets-plugin/internal/diff"
)

func TestComputeEditsIdentical(t *testing.T) {
	edits := ComputeEdits("hello\n", "hello\n")
	if len(edits) != 0 {
		t.Errorf("ComputeEdits() for identical strings = %v, want empty", edits)
	}
}

func TestComputeEditsSimple(t *testing.T) {
	before := "line1\nline2\nline3\n"
	after := "line1\nmodified\nline3\n"
	edits := ComputeEdits(before, after)
	if len(edits) == 0 {
		t.Fatal("ComputeEdits() returned no edits for different strings")
	}
	got, err := diff.Apply(before, edits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got != after {
		t.Errorf("Apply(ComputeEdits()) = %q, want %q", got, after)
	}
}

func TestComputeEditsInsert(t *testing.T) {
	before := "line1\nline3\n"
	after := "line1\nline2\nline3\n"
	edits := ComputeEdits(before, after)
	got, err := diff.Apply(before, edits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got != after {
		t.Errorf("Apply(ComputeEdits()) = %q, want %q", got, after)
	}
}

func TestComputeEditsDelete(t *testing.T) {
	before := "line1\nline2\nline3\n"
	after := "line1\nline3\n"
	edits := ComputeEdits(before, after)
	got, err := diff.Apply(before, edits)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got != after {
		t.Errorf("Apply(ComputeEdits()) = %q, want %q", got, after)
	}
}

func TestComputeEditsEmpty(t *testing.T) {
	edits := ComputeEdits("", "")
	if len(edits) != 0 {
		t.Errorf("ComputeEdits() for empty strings = %v, want empty", edits)
	}
}

func TestComputeEditsToUnified(t *testing.T) {
	before := "line1\nline2\nline3\n"
	after := "line1\nchanged\nline3\n"
	edits := ComputeEdits(before, after)
	result, err := diff.ToUnified("before.txt", "after.txt", before, edits, 0)
	if err != nil {
		t.Fatalf("ToUnified() error = %v", err)
	}
	if result == "" {
		t.Error("ToUnified() returned empty for different strings")
	}
}
