package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Want one and only one argument.")
		os.Exit(0)
	}
	fmt.Println(os.Args[1])
}
