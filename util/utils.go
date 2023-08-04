package util

import (
	"sort"
	"unicode"
)

// func Sanitize(str string) string {
// 	var b strings.Builder
// 	for _, r := range str {
// 		if unicode.IsLetter(r) || unicode.IsDigit(r) {
// 			b.WriteRune(r)
// 		} else if unicode.IsSpace(r) {
// 			b.WriteRune(' ')
// 		} else {
// 			b.WriteRune('_')
// 		}
// 	}
// 	s := b.String()
// 	s = strings.TrimSpace(s)
// 	return s
// }

// func FirstLine(text string) string {
// 	var line string
// 	scanner := bufio.NewScanner(strings.NewReader(text))
// 	for scanner.Scan() {
// 		line = scanner.Text()
// 		if len(line) > 0 {
// 			break
// 		}
// 	}
// 	return line
// }

func IsStringEmpty(str string) bool {
	for _, r := range str {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
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
	if len(s) < 2 {
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

func Contains[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func Union[T comparable](a []T, b []T) []T {
	var result []T = a
	for _, elem := range b {
		if !Contains(a, elem) {
			result = append(result, elem)
		}
	}
	return result
}

func Intersection[T comparable](a []T, b []T) []T {
	var result []T
	for _, elem := range b {
		if Contains(a, elem) {
			result = append(result, elem)
		}
	}
	return result
}

func Exclusion[T comparable](a []T, b []T) []T {
	var result []T
	for _, elem := range a {
		if !Contains(b, elem) {
			result = append(result, elem)
		}
	}
	return result
}
