package library

import (
	"math/big"
	"regexp"
	"strings"
)

// ---------------------------------------------------------------------------
// Numeric string comparison (matching Crystal src/util/numeric_sort.cr)
// ---------------------------------------------------------------------------

var numericSplitRe = regexp.MustCompile(`([^\d\n\r]*)(\d*)([^\d\n\r]*)`)

// isNumeric returns true if the string starts with a digit sequence.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	return s[0] >= '0' && s[0] <= '9'
}

// splitByAlphaNumeric splits a string into alternating non-digit and digit
// groups, matching Crystal's split_by_alphanumeric.
func splitByAlphaNumeric(s string) []string {
	var parts []string
	matches := numericSplitRe.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		// m[1] = non-digits, m[2] = digits, m[3] = non-digits
		if m[1] != "" {
			parts = append(parts, m[1])
		}
		if m[2] != "" {
			parts = append(parts, m[2])
		}
		if m[3] != "" {
			parts = append(parts, m[3])
		}
	}
	return parts
}

// compareNumerically compares two alphanumeric strings using natural sort
// order: numeric segments are compared as numbers (big.Int), string segments
// lexicographically.  Returns -1, 0, or +1.
func compareNumerically(a, b string) int {
	pa := splitByAlphaNumeric(a)
	pb := splitByAlphaNumeric(b)

	// Pad shorter slice with empty strings
	if len(pa) < len(pb) {
		extra := make([]string, len(pb)-len(pa))
		pa = append(pa, extra...)
	} else if len(pb) < len(pa) {
		extra := make([]string, len(pa)-len(pb))
		pb = append(pb, extra...)
	}

	for i := 0; i < len(pa) && i < len(pb); i++ {
		if pa[i] == "" && pb[i] != "" {
			return -1
		}
		if pb[i] == "" && pa[i] != "" {
			return 1
		}
		if pa[i] == "" && pb[i] == "" {
			continue
		}

		aIsNum := isNumeric(pa[i])
		bIsNum := isNumeric(pb[i])

		if aIsNum && bIsNum {
			// Compare as big integers
			ai := new(big.Int)
			bi := new(big.Int)
			ai.SetString(pa[i], 10)
			bi.SetString(pb[i], 10)
			if cmp := ai.Cmp(bi); cmp != 0 {
				return cmp
			}
		} else {
			if cmp := strings.Compare(pa[i], pb[i]); cmp != 0 {
				return cmp
			}
		}
	}

	// All compared equal — shorter slice wins
	if len(pa) < len(pb) {
		return -1
	}
	if len(pa) > len(pb) {
		return 1
	}
	return 0
}

// ---------------------------------------------------------------------------
// ChapterSorter — weighted alphanumeric comparison
// ---------------------------------------------------------------------------

// chapterSorter provides a compare method that gives more weight to
// leading numeric segments for natural chapter ordering.
type chapterSorter struct {
	titles []string
}

func newChapterSorter(titles []string) *chapterSorter {
	return &chapterSorter{titles: titles}
}

// compare returns -1, 0, or +1, following the Crystal ChapterSorter logic
// which weights the comparison so "Chapter 1" < "Chapter 2" < "Chapter 10".
// For Phase 2 it delegates to compareNumerically.
func (cs *chapterSorter) compare(a, b string) int {
	return compareNumerically(a, b)
}
