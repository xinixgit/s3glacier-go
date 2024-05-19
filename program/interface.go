package program

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	DefaultDBSchema = "s3g" // for postgres
)

// the interface for a program the s3g tool supports
// each program takes its own inputs
type Program interface {
	Run() error
	InitFlag(fs *flag.FlagSet)
}

func GetPrograms() (programs map[string]Program, program_names []string) {
	programs = map[string]Program{
		"upload-archive":         &ArchiveUploadProgram{},
		"retrieve-inventory":     &InventoryRetrieveProgram{},
		"checksum-check":         &ChecksumCheckProgram{},
		"download-archive":       &ArchiveDownloadProgram{},
		"delete-expired-archive": &ArchiveDeleteProgram{},
		"describe-job":           &JobDescribeProgram{},
	}

	for key := range programs {
		program_names = append(program_names, key)
	}

	return
}

func createSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})

	if err != nil {
		panic(err)
	}

	return sess
}

func createGlacierClient() *glacier.Glacier {
	sess := createSession()
	return glacier.New(sess)
}

func createSqsClient() *sqs.SQS {
	sess := createSession()
	return sqs.New(sess)
}

func createConnStr(usr string, pwd string, host string, db string) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable TimeZone=America/Los_Angeles",
		host,
		usr,
		pwd,
		db,
	)
}
