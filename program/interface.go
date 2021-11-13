package program

import (
	"flag"
)

type Program interface {
	Run()
	InitFlag(fs *flag.FlagSet)
}

func GetPrograms() (programs map[string]Program, program_names []string) {
	programs = map[string]Program{
		"upload-archive": &UploadArchive{},
		"checksum-check": &ChecksumCheck{},
	}

	for key := range programs {
		program_names = append(program_names, key)
	}

	return
}
