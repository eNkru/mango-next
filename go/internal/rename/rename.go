// Package rename ports Crystal src/rename.cr (Rename::Rule DSL).
package rename

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// VHash is the variable map for Render (Crystal VHash).
type VHash map[string]string

type node interface {
	render(h VHash) string
}

type variable struct {
	id string
}

func (v variable) render(h VHash) string {
	if s, ok := h[v.id]; ok {
		return s
	}
	return ""
}

type pattern struct {
	vars []variable
}

func (p pattern) render(h VHash) string {
	for _, v := range p.vars {
		if _, ok := h[v.id]; ok {
			return v.render(h)
		}
	}
	return ""
}

type group struct {
	parts []node // pattern or literal stringNode
}

type stringNode string

func (s stringNode) render(h VHash) string { return string(s) }

func (g group) render(h VHash) string {
	for _, p := range g.parts {
		if pat, ok := p.(pattern); ok {
			if pat.render(h) == "" {
				return ""
			}
		}
	}
	var b strings.Builder
	for _, p := range g.parts {
		b.WriteString(p.render(h))
	}
	return b.String()
}

// Rule is a parsed rename template.
type Rule struct {
	parts []node
}

// Parse parses a rename rule string (Crystal Rule.new).
func Parse(str string) (*Rule, error) {
	r, err := parse(str)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse rename rule %s. Error: %v", str, err)
	}
	return r, nil
}

func parse(str string) (*Rule, error) {
	rule := &Rule{}
	var chars []rune
	var pat *pattern
	var grp *group

	flushLiteral := func() {
		if len(chars) == 0 {
			return
		}
		s := string(chars)
		chars = nil
		if pat != nil {
			pat.vars = append(pat.vars, variable{id: strings.TrimSpace(s)})
		} else if grp != nil {
			grp.parts = append(grp.parts, stringNode(s))
		} else {
			rule.parts = append(rule.parts, stringNode(s))
		}
	}

	for i, char := range str {
		if strings.ContainsRune("[]{}|", char) && len(chars) > 0 {
			flushLiteral()
		}

		switch char {
		case '[':
			if grp != nil || pat != nil {
				return nil, fmt.Errorf("nested groups are not allowed")
			}
			grp = &group{}
		case ']':
			if grp == nil {
				return nil, fmt.Errorf("unmatched ] at position %d", i)
			}
			if pat != nil {
				return nil, fmt.Errorf("patterns (`{}`) should be closed before closing the group (`[]`)")
			}
			rule.parts = append(rule.parts, *grp)
			grp = nil
		case '{':
			if pat != nil {
				return nil, fmt.Errorf("nested patterns are not allowed")
			}
			pat = &pattern{}
		case '}':
			if pat == nil {
				return nil, fmt.Errorf("unmatched } at position %d", i)
			}
			if grp != nil {
				grp.parts = append(grp.parts, *pat)
			} else {
				rule.parts = append(rule.parts, *pat)
			}
			pat = nil
		case '|':
			if pat == nil {
				chars = append(chars, char)
			}
			// inside pattern: separator only (already flushed literal above)
		default:
			if char == '/' {
				return nil, fmt.Errorf("the character %c at position %d is not allowed", char, i)
			}
			chars = append(chars, char)
		}
		_ = utf8.RuneLen(char)
	}

	if len(chars) > 0 {
		// Crystal: remaining chars pushed as string on rule only (not into open pattern)
		// Actually Crystal pushes to @ary only when chars left after loop — always to @ary
		// not into pattern/group. Re-read:
		// unless chars.empty?; @ary.push chars.join; end
		// So trailing text only goes to rule root, even if pattern/group open — then
		// unclosed checks fire. If pattern open with pending var text, that text is lost
		// to root as string? Crystal pushes entire remaining as one string onto @ary.
		rule.parts = append(rule.parts, stringNode(string(chars)))
		chars = nil
	}
	if pat != nil {
		return nil, fmt.Errorf("unclosed pattern {")
	}
	if grp != nil {
		return nil, fmt.Errorf("unclosed group [")
	}
	return rule, nil
}

// Render applies variables and post-processes the filename.
func (r *Rule) Render(h VHash) string {
	var b strings.Builder
	for _, p := range r.parts {
		b.WriteString(p.render(h))
	}
	return postProcess(strings.TrimSpace(b.String()))
}

var illegal = regexp.MustCompile(`[/?<>\\:*|"^]`)

func postProcess(str string) string {
	if str == ".." {
		return "_"
	}
	str = strings.TrimRight(str, " .")
	return illegal.ReplaceAllString(str, "_")
}
