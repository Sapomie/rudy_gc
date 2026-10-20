package moviex

import "strings"

func splitOrderClause(s string) []string {
	out := make([]string, 0, strings.Count(s, ",")+1)
	start := 0
	parenthesisDepth := 0

	appendPart := func(end int) {
		if item := strings.TrimSpace(s[start:end]); item != "" {
			out = append(out, item)
		}
	}

	for index, char := range s {
		switch char {
		case '(':
			parenthesisDepth++
		case ')':
			if parenthesisDepth > 0 {
				parenthesisDepth--
			}
		case ',':
			if parenthesisDepth == 0 {
				appendPart(index)
				start = index + 1
			}
		}
	}
	appendPart(len(s))
	return out
}
