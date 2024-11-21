package handlers

import "strings"

func splitURL(s string, n int) []string {
	args := make([]string, n)
	for i := range n {
		args[i] = ""
	}

	for i, s := range strings.SplitN(s, "/", n) {
		args[i] = s
	}

	return args
}
