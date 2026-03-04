// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"github.com/norman-abramovitz/cf-targets-plugin/internal/diff/lcs"
)

// Lines computes the differences between two strings, line by line.
// Each line is compared as an atomic unit using the LCS algorithm.
// This is the recommended replacement for myers.ComputeEdits.
func Lines(before, after string) []Edit {
	if before == after {
		return nil
	}
	beforeLines, bOffsets := splitLinesWithOffsets(before)
	afterLines, _ := splitLinesWithOffsets(after)
	diffs := lcs.DiffLines(beforeLines, afterLines)

	res := make([]Edit, len(diffs))
	for i, d := range diffs {
		res[i] = Edit{
			Start: bOffsets[d.Start],
			End:   bOffsets[d.End],
			New:   strings.Join(afterLines[d.ReplStart:d.ReplEnd], ""),
		}
	}
	return res
}

// splitLinesWithOffsets splits text into lines (including trailing \n)
// and returns both the lines and byte offsets of each line start,
// plus a final offset equal to len(text).
func splitLinesWithOffsets(text string) ([]string, []int) {
	var lines []string
	offsets := []int{0}
	start := 0
	for i, r := range text {
		if r == '\n' {
			lines = append(lines, text[start:i+1])
			start = i + 1
			offsets = append(offsets, start)
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
		offsets = append(offsets, len(text))
	}
	return lines, offsets
}

// Strings computes the differences between two strings.
// The resulting edits respect rune boundaries.
func Strings(before, after string) []Edit {
	if before == after {
		return nil // common case
	}

	if isASCII(before) && isASCII(after) {
		// TODO(adonovan): opt: specialize diffASCII for strings.
		return diffASCII([]byte(before), []byte(after))
	}
	return diffRunes([]rune(before), []rune(after))
}

// Bytes computes the differences between two byte slices.
// The resulting edits respect rune boundaries.
func Bytes(before, after []byte) []Edit {
	if bytes.Equal(before, after) {
		return nil // common case
	}

	if isASCII(before) && isASCII(after) {
		return diffASCII(before, after)
	}
	return diffRunes(runes(before), runes(after))
}

func diffASCII(before, after []byte) []Edit {
	diffs := lcs.DiffBytes(before, after)

	// Convert from LCS diffs.
	res := make([]Edit, len(diffs))
	for i, d := range diffs {
		res[i] = Edit{d.Start, d.End, string(after[d.ReplStart:d.ReplEnd])}
	}
	return res
}

func diffRunes(before, after []rune) []Edit {
	diffs := lcs.DiffRunes(before, after)

	// The diffs returned by the lcs package use indexes
	// into whatever slice was passed in.
	// Convert rune offsets to byte offsets.
	res := make([]Edit, len(diffs))
	lastEnd := 0
	utf8Len := 0
	for i, d := range diffs {
		utf8Len += runesLen(before[lastEnd:d.Start]) // text between edits
		start := utf8Len
		utf8Len += runesLen(before[d.Start:d.End]) // text deleted by this edit
		res[i] = Edit{start, utf8Len, string(after[d.ReplStart:d.ReplEnd])}
		lastEnd = d.End
	}
	return res
}

// runes is like []rune(string(bytes)) without the duplicate allocation.
func runes(bytes []byte) []rune {
	n := utf8.RuneCount(bytes)
	runes := make([]rune, n)
	for i := range n {
		r, sz := utf8.DecodeRune(bytes)
		bytes = bytes[sz:]
		runes[i] = r
	}
	return runes
}

// runesLen returns the length in bytes of the UTF-8 encoding of runes.
func runesLen(runes []rune) (len int) {
	for _, r := range runes {
		len += utf8.RuneLen(r)
	}
	return len
}

// isASCII reports whether s contains only ASCII.
func isASCII[S string | []byte](s S) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
