package main

import "fmt"

func foo(a int) bool {
	if a > 3 {
		fmt.Println("big")
		return true
	}
	if a > 2 {
		fmt.Println("big")
		return true
	}
	if a > 1 {
		fmt.Println("big")
		return true
	}
	if false {
		// comment
		a++
	}
	if a > -1 {
		return false
	}
	if a < -10 {
		return true
	}
	return false
}
