// Package plugin contains a proof-of-concept goja-based JS sandbox that mirrors
// the Crystal duktape runtime (src/plugin/plugin.cr). Phase 0 goal: prove goja
// can host the `mango.*` helper functions and run a real v1 plugin. The full
// plugin system (v1/v2, subscriptions, downloader) lands in Phase 3.
package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/dop251/goja"
)

// Sandbox is a minimal goja runtime with the mango helper object injected.
type Sandbox struct {
	vm          *goja.Runtime
	httpClient  *http.Client
	store       *jsStore
	infoDir     string
}

// NewSandbox builds a goja VM and injects the `mango` global, mirroring
// def_helper_functions in plugin.cr. timeoutSeconds <= 0 uses 30s.
func NewSandbox(storagePath, infoDir string, timeoutSeconds int) (*Sandbox, error) {
	store, err := newStore(storagePath)
	if err != nil {
		return nil, fmt.Errorf("init plugin store: %w", err)
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	// mirrors Crystal src/util/proxy.cr: respect HTTP(S)_PROXY / NO_PROXY
	s := &Sandbox{
		vm: goja.New(),
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		store:   store,
		infoDir: infoDir,
	}
	if err := s.installMango(); err != nil {
		return nil, err
	}
	return s, nil
}

// Eval runs JS source and returns the resulting value.
func (s *Sandbox) Eval(src string) (goja.Value, error) {
	return s.vm.RunString(src)
}

func (s *Sandbox) installMango() error {
	mango := s.vm.NewObject()

	// mango.get(url[, headers]) -> {status_code, body, headers}
	mango.Set("get", func(call goja.FunctionCall) goja.Value {
		return s.httpDo(http.MethodGet, call, false)
	})

	// mango.post(url, body[, headers]) -> {status_code, body, headers}
	mango.Set("post", func(call goja.FunctionCall) goja.Value {
		return s.httpDo(http.MethodPost, call, true)
	})

	// mango.css(html, selector) -> [outerHTML, ...]
	mango.Set("css", func(call goja.FunctionCall) goja.Value {
		html := call.Argument(0).String()
		selector := call.Argument(1).String()
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			return s.vm.ToValue([]string{})
		}
		var out []string
		doc.Find(selector).Each(func(_ int, sel *goquery.Selection) {
			if h, err := goquery.OuterHtml(sel); err == nil {
				out = append(out, h)
			}
		})
		if out == nil {
			out = []string{}
		}
		return s.vm.ToValue(out)
	})

	// mango.text(html) -> inner text of first body child, "" on failure
	mango.Set("text", func(call goja.FunctionCall) goja.Value {
		html := call.Argument(0).String()
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			return s.vm.ToValue("")
		}
		first := doc.Find("body").Children().First()
		if first.Length() == 0 {
			return s.vm.ToValue("")
		}
		return s.vm.ToValue(first.Text())
	})

	// mango.attribute(html, name) -> attr value or undefined
	mango.Set("attribute", func(call goja.FunctionCall) goja.Value {
		html := call.Argument(0).String()
		name := call.Argument(1).String()
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			return goja.Undefined()
		}
		first := doc.Find("body").Children().First()
		if first.Length() == 0 {
			return goja.Undefined()
		}
		if v, ok := first.Attr(name); ok {
			return s.vm.ToValue(v)
		}
		return goja.Undefined()
	})

	// mango.raise(msg) -> throws a JS exception carrying msg
	mango.Set("raise", func(call goja.FunctionCall) goja.Value {
		msg := call.Argument(0).String()
		panic(s.vm.ToValue(fmt.Sprintf("PluginException: %s", msg)))
	})

	// mango.storage(key[, value]) -> string or undefined
	mango.Set("storage", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		if len(call.Arguments) == 2 {
			val := call.Argument(1).String()
			if err := s.store.set(key, val); err != nil {
				panic(s.vm.ToValue("StorageError: " + err.Error()))
			}
			return goja.Undefined()
		}
		if val, ok := s.store.get(key); ok {
			return s.vm.ToValue(val)
		}
		return goja.Undefined()
	})

	// mango.settings(key) -> string or undefined (v2+ only)
	mango.Set("settings", func(call goja.FunctionCall) goja.Value {
		key := call.Argument(0).String()
		val := readPluginSetting(s.infoDir, key)
		if val == "" {
			return goja.Undefined()
		}
		return s.vm.ToValue(val)
	})

	return s.vm.Set("mango", mango)
}

// readPluginSetting reads a settings value from info.json in the plugin directory.
func readPluginSetting(infoDir, key string) string {
	infoPath := infoDir + "/info.json"
	raw, err := os.ReadFile(infoPath)
	if err != nil {
		return ""
	}
	var info struct {
		Settings map[string]string `json:"settings"`
	}
	if err := json.Unmarshal(raw, &info); err != nil {
		return ""
	}
	return info.Settings[key]
}

func (s *Sandbox) httpDo(method string, call goja.FunctionCall, hasBody bool) goja.Value {
	url := call.Argument(0).String()
	var bodyReader io.Reader
	headerArgIdx := 1
	if hasBody {
		bodyReader = strings.NewReader(call.Argument(1).String())
		headerArgIdx = 2
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		panic(s.vm.ToValue(err.Error()))
	}
	if h := call.Argument(headerArgIdx); h != nil && !goja.IsUndefined(h) && !goja.IsNull(h) {
		if obj := h.ToObject(s.vm); obj != nil {
			for _, k := range obj.Keys() {
				req.Header.Add(k, obj.Get(k).String())
			}
		}
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		panic(s.vm.ToValue(err.Error()))
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	headers := map[string]string{}
	for k, v := range resp.Header {
		headers[k] = strings.Join(v, ",")
	}

	result := map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(body),
		"headers":     headers,
	}
	return s.vm.ToValue(result)
}
