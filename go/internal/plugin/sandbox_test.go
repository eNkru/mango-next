package plugin

import (
	"reflect"
	"testing"

	"github.com/dop251/goja"
)

// These cases are transcribed 1:1 from spec/plugin_spec.cr to prove goja
// produces the same results as the duktape+myhtml runtime.

func newTestSandbox(t *testing.T) *Sandbox {
	t.Helper()
	s, err := NewSandbox("", "")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func evalString(t *testing.T, s *Sandbox, src string) string {
	t.Helper()
	v, err := s.Eval(src)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	return v.String()
}

func TestMangoText(t *testing.T) {
	s := newTestSandbox(t)
	got := evalString(t, s, `mango.text('<a href="https://github.com">Click Me<a>');`)
	if got != "Click Me" {
		t.Errorf("mango.text = %q, want %q", got, "Click Me")
	}
}

func TestMangoTextEmpty(t *testing.T) {
	s := newTestSandbox(t)
	got := evalString(t, s, `mango.text('<img src="https://github.com" />');`)
	if got != "" {
		t.Errorf("mango.text = %q, want empty", got)
	}
}

func TestMangoCSS(t *testing.T) {
	s := newTestSandbox(t)
	v, err := s.Eval(`mango.css('<ul><li class="test">A</li><li class="test">B</li><li>C</li></ul>', 'li.test');`)
	if err != nil {
		t.Fatal(err)
	}
	got := v.Export()
	want := []string{`<li class="test">A</li>`, `<li class="test">B</li>`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mango.css = %#v, want %#v", got, want)
	}
}

func TestMangoCSSNoMatch(t *testing.T) {
	s := newTestSandbox(t)
	v, err := s.Eval(`mango.css('<ul><li class="test">A</li></ul>', 'li.noclass');`)
	if err != nil {
		t.Fatal(err)
	}
	arr := v.Export()
	rv := reflect.ValueOf(arr)
	if rv.Kind() != reflect.Slice || rv.Len() != 0 {
		t.Errorf("mango.css no-match = %#v, want empty slice", arr)
	}
}

func TestMangoAttribute(t *testing.T) {
	s := newTestSandbox(t)
	got := evalString(t, s, `mango.attribute('<a href="https://github.com">Click Me<a>', 'href');`)
	if got != "https://github.com" {
		t.Errorf("mango.attribute = %q, want https://github.com", got)
	}
}

func TestMangoAttributeUndefined(t *testing.T) {
	s := newTestSandbox(t)
	v, err := s.Eval(`mango.attribute('<div />', 'href') === undefined;`)
	if err != nil {
		t.Fatal(err)
	}
	if !v.ToBoolean() {
		t.Error("mango.attribute no-match should be undefined")
	}
}

// https://github.com/hkalexling/Mango/issues/320
func TestMangoAttributeTagsInValue(t *testing.T) {
	s := newTestSandbox(t)
	got := evalString(t, s, `mango.attribute('<div data-a="<img />" data-b="test" />', 'data-b');`)
	if got != "test" {
		t.Errorf("mango.attribute = %q, want test", got)
	}
}

// Proves a v1-style plugin flow (define + call functions returning JSON) works.
func TestV1PluginFlow(t *testing.T) {
	s := newTestSandbox(t)
	_, err := s.Eval(`
		function listChapters(query) {
			return JSON.stringify({
				title: "Series " + query,
				chapters: [{ id: "c1", title: "Chapter 1" }]
			});
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	v, err := s.Eval(`listChapters('abc')`)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := goja.AssertFunction(v); ok {
		t.Fatal("expected string result")
	}
	if v.String() == "" {
		t.Error("listChapters returned empty")
	}
}
