package rename

import (
	"strings"
	"testing"
)

func TestParseErrors(t *testing.T) {
	cases := []string{"[[]]", "{{}}", "[", "{", "[{]}", "]", "[}", "hello/world"}
	for _, c := range cases {
		if _, err := Parse(c); err == nil {
			t.Errorf("Parse(%q) expected error", c)
		}
	}
}

func TestPatternPipe(t *testing.T) {
	rule, err := Parse("{a|b|c}")
	if err != nil {
		t.Fatal(err)
	}
	if got := rule.Render(VHash{"b": "b"}); got != "b" {
		t.Fatalf("got %q", got)
	}
	if got := rule.Render(VHash{"a": "a", "b": "b"}); got != "a" {
		t.Fatalf("got %q", got)
	}
}

func TestSpacesInPatterns(t *testing.T) {
	rule, err := Parse("{  a }")
	if err != nil {
		t.Fatal(err)
	}
	if got := rule.Render(VHash{"a": "a"}); got != "a" {
		t.Fatalf("got %q", got)
	}
}

func TestStripOuterSpaces(t *testing.T) {
	rule, err := Parse("  hello ")
	if err != nil {
		t.Fatal(err)
	}
	if got := rule.Render(VHash{"a": "a"}); got != "hello" {
		t.Fatalf("got %q", got)
	}
}

func TestExamples(t *testing.T) {
	rule, err := Parse("[Ch. {chapter }] {title | id} testing")
	if err != nil {
		t.Fatal(err)
	}
	if got := rule.Render(VHash{"id": "ID"}); got != "ID testing" {
		t.Fatalf("got %q", got)
	}
	if got := rule.Render(VHash{"chapter": "CH", "id": "ID"}); got != "Ch. CH ID testing" {
		t.Fatalf("got %q", got)
	}
	if got := rule.Render(VHash{}); got != "testing" {
		t.Fatalf("got %q", got)
	}
}

func TestIllegalChars(t *testing.T) {
	rule, err := Parse("{a}")
	if err != nil {
		t.Fatal(err)
	}
	got := rule.Render(VHash{"a": `/?<>:*|"^`})
	if got != "_________" {
		t.Fatalf("got %q", got)
	}
}

func TestStripTrailing(t *testing.T) {
	rule, err := Parse("hello. world. ..")
	if err != nil {
		t.Fatal(err)
	}
	if got := rule.Render(VHash{}); got != "hello. world" {
		t.Fatalf("got %q", got)
	}
}

func TestErrorMessageWraps(t *testing.T) {
	_, err := Parse("[")
	if err == nil || !strings.Contains(err.Error(), "Failed to parse rename rule") {
		t.Fatalf("expected wrap error, got %v", err)
	}
}
