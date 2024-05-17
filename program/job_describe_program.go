package program

import (
	"flag"
	"fmt"
	"s3glacier-go/adapter"
)

type JobDescribeProgram struct {
	jobID string
	vault string
}

func (p *JobDescribeProgram) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.jobID, "id", "", "The id of the job")
	fs.StringVar(&p.vault, "v", "", "The name of the vault the job is for")
}

func (p *JobDescribeProgram) Run() {
	s3g := createGlacierClient()
	csp := adapter.NewCloudServiceProvider(s3g)

	jd, err := csp.DescribeJob(&p.jobID, &p.vault)
	if err != nil {
		fmt.Println("Failed to get job status")
		panic(err)
	}

	fmt.Println(*jd)
}
