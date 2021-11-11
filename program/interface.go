package program

type Program interface{ Run() }

func GetPrograms() (programs map[string]Program, program_names []string) {
	programs = map[string]Program{
		"upload-archive": &UploadArchive{},
	}

	for key := range programs {
		program_names = append(program_names, key)
	}

	return
}
