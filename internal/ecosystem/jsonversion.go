package ecosystem

import "fmt"

// replaceTopLevelJSONVersion replaces the value of the top-level "version" key
// in a JSON or JSONC document, preserving all original formatting, key order,
// and comments. Unlike a plain regex replacement, it tracks nesting depth so
// nested "version" fields (e.g. inside dependencies or tool config) are left
// untouched. Returns an error if no top-level string-valued "version" key exists.
func replaceTopLevelJSONVersion(content []byte, newVersion string) ([]byte, error) {
	depth := 0
	i := 0
	n := len(content)

	// skipInsignificant advances past whitespace and JSONC comments starting at
	// position p and returns the index of the next significant byte.
	skipInsignificant := func(p int) int {
		for p < n {
			switch {
			case content[p] == ' ' || content[p] == '\t' || content[p] == '\n' || content[p] == '\r':
				p++
			case content[p] == '/' && p+1 < n && content[p+1] == '/':
				p += 2
				for p < n && content[p] != '\n' {
					p++
				}
			case content[p] == '/' && p+1 < n && content[p+1] == '*':
				p += 2
				for p+1 < n && !(content[p] == '*' && content[p+1] == '/') {
					p++
				}
				p += 2
			default:
				return p
			}
		}
		return p
	}

	// scanString returns the index just past the closing quote of the string
	// starting at the opening quote position p.
	scanString := func(p int) int {
		p++ // opening quote
		for p < n {
			if content[p] == '\\' {
				p += 2
				continue
			}
			if content[p] == '"' {
				return p + 1
			}
			p++
		}
		return p
	}

	for i < n {
		i = skipInsignificant(i)
		if i >= n {
			break
		}
		switch content[i] {
		case '{', '[':
			depth++
			i++
		case '}', ']':
			depth--
			i++
		case '"':
			strStart := i
			strEnd := scanString(i)
			i = strEnd
			if depth != 1 {
				continue
			}
			// A string at depth 1 followed by ':' is a top-level key.
			next := skipInsignificant(i)
			if next >= n || content[next] != ':' {
				continue
			}
			key := string(content[strStart+1 : strEnd-1])
			i = next + 1
			if key != "version" {
				continue
			}
			valStart := skipInsignificant(i)
			if valStart >= n || content[valStart] != '"' {
				return nil, fmt.Errorf("top-level version field is not a string")
			}
			valEnd := scanString(valStart)
			if valEnd > n || content[valEnd-1] != '"' {
				return nil, fmt.Errorf("unterminated version string")
			}
			result := make([]byte, 0, n+len(newVersion))
			result = append(result, content[:valStart+1]...)
			result = append(result, newVersion...)
			result = append(result, content[valEnd-1:]...)
			return result, nil
		default:
			i++
		}
	}

	return nil, fmt.Errorf("no version field found")
}
