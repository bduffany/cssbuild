package cssbuild

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/bduffany/cssbuild/cssbuild/css"
	"github.com/tdewolff/parse/v2"
)

const (
	debug = false

	local scopeType = iota
	global

	randSuffixLength = 8
	randSuffixChars  = "abcdefghijklmnopqrstuvwxyz0123456789"
)

type scopeType int

type TransformOpts struct {
	Suffix []byte
}

// Transform reads a module stylesheet from the given reader, and writes the
// transformed stylesheet to the given writer.
func Transform(r io.Reader, w io.Writer, opts *TransformOpts) error {
	var debugBuf bytes.Buffer
	if debug {
		w = io.MultiWriter(w, &debugBuf)
	}

	{
		optsCopy := *opts
		opts = &optsCopy
	}
	if len(opts.Suffix) == 0 {
		opts.Suffix = RandomSuffix()
	}

	p := css.NewParser(parse.NewInput(r), false /*=inline*/)

	var buf []byte
	var indent []byte
	blockScope := local
	for {
		// Consume the next token.
		gt, tt, text := p.Next()
		err := p.Err()
		values := p.Values()

		if debug {
			fmt.Printf("\033[90m#%s: tt=%s, text=%q, values=%v\033[0m\n", gt, tt, string(text), values)
		}

		if err == io.EOF {
			if len(buf) > 0 {
				return fmt.Errorf("unexpected: unflushed contents %q when seeing EOF", string(buf))
			}
			return nil
		}
		// Return non-EOF errors immediately.
		// EOF may or may not be an error depending on the current parse state
		// (decided by the state machine below).
		if err != nil && err != io.EOF {
			// TODO: include line number & filename in error
			return err
		}

		if gt == css.QualifiedRuleGrammar || gt == css.BeginRulesetGrammar {
			b, endScope := transformSelector(text, values, opts)
			buf = append(buf, b...)
			blockScope = endScope
			if gt == css.BeginRulesetGrammar {
				if len(buf) > 0 && buf[len(buf)-1] != ' ' {
					buf = append(buf, ' ')
				}
				buf = append(buf, '{')
			}
			if gt == css.QualifiedRuleGrammar {
				buf = append(buf, ',')
			}
		} else if gt == css.BeginAtRuleGrammar {
			buf = append(buf, transformAtRule(text, values, opts)...)
			buf = append(buf, ' ', '{')
		} else {
			if gt == css.CustomPropertyGrammar {
				buf = append(buf, text...)
				buf = append(buf, ':')
			}
			if gt == css.DeclarationGrammar {
				buf = append(buf, text...)
				buf = append(buf, ':', ' ')
			}

			if gt == css.DeclarationGrammar && string(text) == "animation" {
				buf = append(buf, transformAnimationProperty(values, blockScope, opts)...)
			} else if gt == css.DeclarationGrammar && string(text) == "animation-name" {
				buf = append(buf, transformAnimationNameProperty(values, blockScope, opts)...)
			} else if gt != css.EndAtRuleGrammar && gt != css.EndRulesetGrammar {
				for _, val := range p.Values() {
					buf = append(buf, val.Data...)
				}
			}
			if gt == css.BeginAtRuleGrammar || gt == css.BeginRulesetGrammar {
				buf = append(buf, '{')
			} else if gt == css.AtRuleGrammar || gt == css.DeclarationGrammar || gt == css.CustomPropertyGrammar {
				buf = append(buf, ';')
			} else if gt == css.EndAtRuleGrammar || gt == css.EndRulesetGrammar {
				if len(indent) >= 2 {
					indent = indent[:len(indent)-2]
				}
				buf = append(buf, '}')
				blockScope = local
			}
		}

		if len(buf) == 1 && buf[0] == '}' && len(indent) == 0 {
			// Double-newline at top-level rule sets
			// TODO: Double-newline for nested rulesets, unless they're the last
			// one in the ruleset.
			buf = append(buf, '\n')
		}
		buf = append(buf, '\n')

		if len(indent) > 0 {
			if _, err := w.Write(indent); err != nil {
				return err
			}
		}
		if len(buf) > 0 {
			if _, err := w.Write(buf); err != nil {
				return err
			}
			buf = nil
		}

		if debug {
			str := debugBuf.String()
			if str != "" {
				fmt.Printf("%s", debugBuf.String())
				if str[len(str)-1] != '\n' {
					fmt.Printf("\033[90m\t!\\n\033[0m\n")
				}
			}
			debugBuf.Reset()
		}

		if gt == css.BeginAtRuleGrammar || gt == css.BeginRulesetGrammar {
			indent = append(indent, ' ', ' ')
		}
	}
}

func RandomSuffix() []byte {
	out := []byte{'_'}
	s := rand.NewSource(int64(time.Now().UnixNano()))
	r := rand.New(s)
	for i := 0; i < randSuffixLength; i++ {
		index := r.Intn(len(randSuffixChars))
		out = append(out, randSuffixChars[index])
	}
	return out
}

func transformSelector(text []byte, values []css.Token, opts *TransformOpts) (buf []byte, endScope scopeType) {
	scopeMode := local
	scopeStack := []scopeType{}

	isClassName := false
	funcStack := []string{}
	skip := 0
	for i := 0; i < len(values); i++ {
		val := values[i]

		if val.TokenType == css.ColonToken && i+1 < len(values) {
			// :local, :global, :local(, :global( should not be written to output

			next := values[i+1]
			nextStr := string(next.Data)
			if (next.TokenType == css.FunctionToken || next.TokenType == css.IdentToken) &&
				(nextStr == "local(" || nextStr == "global(" || nextStr == "local" || nextStr == "global") {
				skip += 2
				// If there's another whitespace after the ":local" / ":global" mode
				// selector, skip that too.
				if i+2 < len(values) {
					nextNext := values[i+2]
					if nextNext.TokenType == css.WhitespaceToken {
						skip++
					}
				}

				// Hack: pre-emptively set the scope mode before `i` actually reaches
				// the next token. This simplifies impl.
				if nextStr == "local" {
					scopeMode = local
				} else if nextStr == "global" {
					scopeMode = global
				}
			}
		} else if val.TokenType == css.FunctionToken {
			// Push funcs to stack
			funcStack = append(funcStack, string(val.Data))
			if string(val.Data) == "local(" {
				scopeStack = append(scopeStack, local)
			} else if string(val.Data) == "global(" {
				scopeStack = append(scopeStack, global)
			}
		} else if val.TokenType == css.RightParenthesisToken {
			// Pop from func stack when seeing closing parens
			if len(funcStack) > 0 {
				popped := funcStack[len(funcStack)-1]
				funcStack = funcStack[:len(funcStack)-1]
				if popped == "local(" || popped == "global(" {
					scopeStack = scopeStack[:len(scopeStack)-1]
					skip++
				}
			}
		}

		if skip == 0 {
			buf = append(buf, val.Data...)
			scope := scopeMode
			if len(scopeStack) > 0 {
				scope = scopeStack[len(scopeStack)-1]
			}
			if isClassName && scope == local {
				buf = append(buf, opts.Suffix...)
			}
		} else {
			skip--
		}
		isClassName = (val.TokenType == css.DelimToken && len(val.Data) == 1 && val.Data[0] == '.')
	}
	return buf, scopeMode
}

func transformAtRule(text []byte, values []css.Token, opts *TransformOpts) (buf []byte) {
	buf = append(buf, text...)
	if string(text) == "@keyframes" {
		// When using @keyframes :global, this confuses the parser and it doesn't
		// add whitespace after @keyframes. Add it here.
		if len(values) > 0 && values[0].TokenType != css.WhitespaceToken {
			buf = append(buf, ' ')
		}
		scope := local
		for i, val := range values {
			if val.TokenType == css.ColonToken && i+1 < len(values) {
				next := values[i+1]
				nextStr := string(next.Data)
				if next.TokenType == css.FunctionToken && nextStr == "global(" || next.TokenType == css.IdentToken && nextStr == "global" {
					scope = global
				}
			}
			if val.TokenType != css.ColonToken && val.TokenType != css.FunctionToken && val.TokenType != css.RightParenthesisToken {
				buf = append(buf, val.Data...)
			}
			if val.TokenType == css.IdentToken && scope != global {
				buf = append(buf, opts.Suffix...)
			}
		}
	} else {
		for _, val := range values {
			buf = append(buf, val.Data...)
		}
	}
	return
}

func transformAnimationProperty(values []css.Token, scope scopeType, opts *TransformOpts) (buf []byte) {
	if scope == global {
		for _, val := range values {
			buf = append(buf, val.Data...)
		}
		return
	}
	// Reference: https://www.w3.org/TR/css-animations-1/#animation

	// local scope
	inFunction := false
	sawTimingFunction := false
	sawIterationCount := false
	sawDirection := false
	sawFillMode := false
	sawPlayState := false

	for _, val := range values {
		buf = append(buf, val.Data...)
		// Parser throws away whitespace after commas; recover it.
		if val.TokenType == css.CommaToken {
			buf = append(buf, ' ')
		}
		if val.TokenType == css.RightParenthesisToken {
			inFunction = false
			continue
		}
		if inFunction || val.TokenType == css.WhitespaceToken || val.TokenType == css.DimensionToken {
			continue
		}
		// Consume next value in list
		if val.TokenType == css.CommaToken {
			inFunction = false
			sawTimingFunction = false
			sawIterationCount = false
			sawDirection = false
			sawFillMode = false
			sawPlayState = false
			continue
		}
		// Consume timing function
		if val.TokenType == css.FunctionToken {
			inFunction = true
			sawTimingFunction = true
			continue
		}
		str := string(val.Data)
		if !sawTimingFunction && (str == "linear" || str == "ease" || str == "ease-in" || str == "ease-out" ||
			str == "ease-in-out" || str == "step-start" || str == "step-end") {
			sawTimingFunction = true
			continue
		}
		if !sawIterationCount && (str == "infinite" || val.TokenType == css.NumberToken) {
			sawIterationCount = true
			continue
		}
		if !sawDirection && (str == "normal" || str == "reverse" || str == "alternate" || str == "alternate-reverse") {
			sawDirection = true
			continue
		}
		if !sawFillMode && (str == "none" || str == "forwards" || str == "backwards" || str == "both") {
			sawFillMode = true
			continue
		}
		if !sawPlayState && (str == "running" || str == "paused") {
			sawPlayState = true
			continue
		}
		// If we see an identifier that can't be parsed as any other property,
		// interpret it as the animation name and apply the suffix.
		if val.TokenType == css.IdentToken {
			buf = append(buf, opts.Suffix...)
		}
	}

	return
}

func transformAnimationNameProperty(values []css.Token, scope scopeType, opts *TransformOpts) (buf []byte) {
	if scope == global {
		for _, val := range values {
			buf = append(buf, val.Data...)
		}
		return
	}

	for _, val := range values {
		buf = append(buf, val.Data...)
		if val.TokenType == css.IdentToken {
			buf = append(buf, opts.Suffix...)
		}
	}

	return
}
