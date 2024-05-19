package program

import (
	"flag"
	"fmt"
	"s3glacier-go/adapter"
	"s3glacier-go/svc"
)

type JobDescribeProgram struct {
	jobID string
	vault string
}

func (p *JobDescribeProgram) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.jobID, "id", "", "The id of the job")
	fs.StringVar(&p.vault, "v", "", "The name of the vault the job is for")
}

func (p *JobDescribeProgram) Run() error {
	s3g := createGlacierClient()
	csp := adapter.NewCloudServiceProvider(s3g)

	jd, err := csp.DescribeJob(&p.jobID, &p.vault)
	if err != nil {
		return fmt.Errorf("failed to get job status: %w", err)
	}

	fmt.Println(*jd)
	fmt.Println()

	if *jd.Completed {
		output, err := csp.GetJobOutput(&p.jobID, &p.vault)
		if err != nil {
			return fmt.Errorf("failed to get job output: %w", err)
		}

		s, err := svc.ReadAllFromStream(output.Body)
		if err != nil {
			return fmt.Errorf("failed to read from job output stream: %w", err)
		}

		fmt.Println(*s)
	}

	return nil
}
