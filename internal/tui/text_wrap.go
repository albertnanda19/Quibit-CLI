package tui

import (
	"strings"
	"unicode/utf8"
)

func wrapSoft(text string, max int) []string {
	text = strings.TrimRight(text, "\r\n")
	if strings.TrimSpace(text) == "" {
		return []string{""}
	}
	if max <= 0 {
		return []string{text}
	}

	paras := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(paras))
	for _, p := range paras {
		p = strings.TrimRight(p, " \t")
		if p == "" {
			out = append(out, "")
			continue
		}
		out = append(out, wrapOneLine(p, max)...)
	}
	return out
}

func wrapOneLine(text string, max int) []string {
	if max <= 0 {
		return []string{text}
	}
	if utf8.RuneCountInString(text) <= max {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	lines := []string{}
	var cur strings.Builder
	curLen := 0
	flush := func() {
		if curLen == 0 {
			return
		}
		lines = append(lines, cur.String())
		cur.Reset()
		curLen = 0
	}

	for _, w := range words {
		wLen := utf8.RuneCountInString(w)
		if curLen == 0 {
			if wLen <= max {
				cur.WriteString(w)
				curLen = wLen
				continue
			}
			rs := []rune(w)
			for len(rs) > 0 {
				chunk := rs
				if len(chunk) > max {
					chunk = rs[:max]
				}
				lines = append(lines, string(chunk))
				rs = rs[len(chunk):]
			}
			continue
		}

		if curLen+1+wLen <= max {
			cur.WriteString(" ")
			cur.WriteString(w)
			curLen += 1 + wLen
			continue
		}

		flush()
		if wLen <= max {
			cur.WriteString(w)
			curLen = wLen
			continue
		}
		rs := []rune(w)
		for len(rs) > 0 {
			chunk := rs
			if len(chunk) > max {
				chunk = rs[:max]
			}
			lines = append(lines, string(chunk))
			rs = rs[len(chunk):]
		}
	}
	flush()
	return lines
}
