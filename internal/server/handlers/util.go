package handlers

import "strings"

func splitURI(s string, n int) []string {
	args := make([]string, n)
	for i := range n {
		args[i] = ""
	}

	for i, s := range strings.SplitN(s, "/", n) {
		args[i] = s
	}

	return args
}
