package util

import (
	"bufio"
	"sort"
	"strings"
	"unicode"
)

func Sanitize(str string) string {
	var b strings.Builder
	for _, r := range str {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) {
			b.WriteRune(' ')
		} else {
			b.WriteRune('_')
		}
	}
	s := b.String()
	s = strings.TrimSpace(s)
	return s
}

func FirstLine(text string) string {
	var line string
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line = scanner.Text()
		if len(line) > 0 {
			break
		}
	}
	return line
}

/*
func RemoveDuplicates[T string | int](sliceList []T) []T {
    allKeys := make(map[T]bool)
    list := []T{}
    for _, item := range sliceList {
        if _, value := allKeys[item]; !value {
            allKeys[item] = true
            list = append(list, item)
        }
    }
    return list
}
*/

func RemoveDuplicateStrings(s []string) []string {
	if len(s) < 1 {
		return s
	}

	sort.Strings(s)
	prev := 1
	for curr := 1; curr < len(s); curr++ {
		if s[curr-1] != s[curr] {
			s[prev] = s[curr]
			prev++
		}
	}

	return s[:prev]
}
