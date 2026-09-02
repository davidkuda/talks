package main

import (
	"fmt"
)

func prettyPrintURL(url string) {
	previous := 0
	for i := range url {
		if url[i] == '?' || url[i] == '&' {
			fmt.Println(url[previous:i])
			previous = i
		}
	}
	fmt.Println(url[previous:])
}

func prettyPrintURLAsTable(url string) {
	rawArgs := make([]string, 0)
	previous := 0
	for i := range url {
		if url[i] == '?' {
			previous = i
			break
		}
	}

	fmt.Println(url[0:previous])

	for i := range url {
		if url[i] == '&' {
			rawArgs = append(rawArgs, url[previous:i])
			previous = i
		}
	}
	rawArgs = append(rawArgs, url[previous:])

	type arg struct{ key, value string }
	args := make([]arg, 0)

	for _, rarg := range rawArgs {
		for i := range rarg {
			if rarg[i] == '=' {
				a := arg{
					key:   rarg[1:i],
					value: rarg[i+1:],
				}
				args = append(args, a)
			}
		}
	}

	var longestKeyLen int
	for _, a := range args {
		if len(a.key) > longestKeyLen {
			longestKeyLen = len(a.key)
		}
	}

	for _, a := range args {
		fmt.Printf("    %-21s\t%s\n", a.key, a.value)
	}
}
