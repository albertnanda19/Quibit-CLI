package tui

func stripANSI(s string) string {
	out := make([]rune, 0, len(s))
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		if rs[i] != 0x1b {
			out = append(out, rs[i])
			continue
		}
		if i+1 < len(rs) && rs[i+1] == '[' {
			i += 2
			for i < len(rs) {
				if (rs[i] >= 'A' && rs[i] <= 'Z') || (rs[i] >= 'a' && rs[i] <= 'z') {
					break
				}
				i++
			}
			continue
		}
	}
	return string(out)
}
