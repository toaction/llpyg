package pysig

import (
	"strings"
)

type Arg struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	DefVal   string `json:"defVal"`
	Optional bool   `json:"optional"`
}

func Parse(sig string) (args []*Arg) {
	sig = strings.TrimSpace(sig)
	sig = strings.TrimPrefix(sig, "(")

	for len(sig) > 0 && sig[0] != ')' {
		// Skip leading whitespace and commas
		sig = strings.TrimLeft(sig, " \t,")
		if len(sig) == 0 || sig[0] == ')' {
			break
		}

		// optional parameters [param1, param2, ...]
		if sig[0] == '[' {
			bracketEnd := findMatchingBracket(sig, '[', ']')
			if bracketEnd > 0 {
				bracketContent := sig[1:bracketEnd]
				bracketContent = "(" + bracketContent + ")"
				optArgs := Parse(bracketContent)
				for _, arg := range optArgs {
					arg.Optional = true
				}
				args = append(args, optArgs...)
				sig = sig[bracketEnd+1:]
				continue
			}
		}

		// complex parameters (a1, a2, ...)
		if sig[0] == '(' {
			parenEnd := findMatchingBracket(sig, '(', ')')
			if parenEnd > 0 {
				paramName := strings.TrimSpace(sig[:parenEnd+1])
				arg := &Arg{Name: paramName}
				args = append(args, arg)
				sig = sig[parenEnd+1:]

				// Check for assignment after the parentheses
				sig = strings.TrimLeft(sig, " \t")
				if len(sig) > 0 && sig[0] == '=' {
					sig = sig[1:]
					arg.DefVal, sig = parseDefVal(sig)
				}
				continue
			}
		}

		// Normal parameter parsing
		pos := strings.IndexAny(sig, ",:=)[")
		if pos <= 0 {
			// Handle case where we reach end of string
			if len(sig) > 0 {
				name := strings.TrimSpace(sig)
				if name != "" && name != ")" {
					args = append(args, &Arg{Name: name})
				}
			}
			return
		}

		argName := strings.TrimSpace(sig[:pos])
		if argName == "" {
			sig = sig[1:]
			continue
		}

		if strings.TrimSpace(argName) == "..." {
			argName = "**kwargs"
		}

		arg := &Arg{Name: argName}
		args = append(args, arg)
		c := sig[pos]
		sig = sig[pos+1:]

		switch c {
		case ',':
			continue
		case ':':
			arg.Type, sig = parseType(sig)
			if strings.HasPrefix(strings.TrimLeft(sig, " \t"), "=") {
				sig = strings.TrimLeft(sig, " \t")
				arg.DefVal, sig = parseDefVal(sig[1:])
			}
		case '=':
			arg.DefVal, sig = parseDefVal(sig)
		case ')':
			return
		case '[':
			// Backtrack - this [ should be handled at the beginning of the loop
			sig = "[" + sig
			continue
		}
	}
	return
}

const (
	allSpecials = "([<'\""
)

var pairStops = map[byte]string{
	'(':  ")" + allSpecials,
	'[':  "]" + allSpecials,
	'<':  ">" + allSpecials,
	'\'': "'" + allSpecials,
	'"':  "\"",
}

func parseText(sig string, stops string) (left string) {
	for {
		pos := strings.IndexAny(sig, stops)
		if pos < 0 {
			return sig
		}
		c := sig[pos]
		if c != stops[0] {
			if pstop, ok := pairStops[c]; ok {
				sig = strings.TrimPrefix(parseText(sig[pos+1:], pstop), pstop[:1])
				continue
			}
		}
		return sig[pos:]
	}
}

func parseDefValText(sig string, stops string) (left string) {
	for {
		pos := strings.IndexAny(sig, stops)
		if pos < 0 {
			return sig
		}
		c := sig[pos]
		// Special handling for '[' when parsing default values
		if c == '[' && strings.Contains(stops, "[") {
			return sig[pos:]
		}
		if c != stops[0] {
			if pstop, ok := pairStops[c]; ok {
				sig = strings.TrimPrefix(parseDefValText(sig[pos+1:], pstop), pstop[:1])
				continue
			}
		}
		return sig[pos:]
	}
}

// stops: "=,)"
func parseType(sig string) (string, string) {
	left := parseText(sig, "=,)"+allSpecials)
	return resultOf(sig, left), left
}

// stops: ",)"
func parseDefVal(sig string) (string, string) {
	left := parseDefValText(sig, ",)["+allSpecials)
	return resultOf(sig, left), left
}

func resultOf(sig, left string) string {
	return strings.TrimSpace(sig[:len(sig)-len(left)])
}

// findMatchingBracket finds the matching closing bracket for the opening bracket at position 0
func findMatchingBracket(sig string, open, close byte) int {
	if len(sig) == 0 || sig[0] != open {
		return -1
	}

	count := 1
	for i := 1; i < len(sig); i++ {
		switch sig[i] {
		case open:
			count++
		case close:
			count--
			if count == 0 {
				return i
			}
		}
	}
	return -1
}
