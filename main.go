package main

import (
	"os"
	"fmt"
	"xddd/s3glacier/program"
)

func main() {
	programs, program_names := program.GetPrograms()

	if len(os.Args) < 2 {
		fmt.Printf("Specify a program to run. Available ones are: %s\n", program_names)
		return
	}

	p := os.Args[1]
	if _, ok := programs[p]; !ok {
		fmt.Printf("Specified program %s not found. Available ones are: %s\n", p, program_names)
		return
	}

	programs[p].Run()
}
