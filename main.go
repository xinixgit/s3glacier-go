package main

import (
	"flag"
	"fmt"
	"os"
	"s3glacier-go/program"
)

var p program.Program

func init() {
	programs, program_names := program.GetPrograms()

	if len(os.Args) < 2 {
		panic(fmt.Sprintf("Specify a program to run. Available ones are: %s\n", program_names))
	}

	pName := os.Args[1]
	if _, ok := programs[pName]; !ok {
		panic(fmt.Sprintf("Specified program %s not found. Available ones are: %s\n", pName, program_names))
	}

	p = programs[pName]

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	p.InitFlag(fs)
	fs.Parse(os.Args[2:])

	fs.VisitAll(func(f *flag.Flag) {
		if f.Value.String() == "" {
			fmt.Printf("Usage: s3glacier %s\n", os.Args[1])
			fs.PrintDefaults()
			panic("End execution now.")
		}
	})
}

func main() {
	p.Run()
}
