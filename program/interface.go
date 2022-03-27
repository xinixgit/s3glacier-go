package program

import (
	"flag"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Program interface {
	Run()
	InitFlag(fs *flag.FlagSet)
}

func GetPrograms() (programs map[string]Program, program_names []string) {
	programs = map[string]Program{
		"upload-archive":     &UploadArchive{},
		"retrieve-inventory": &InventoryRetrieval{},
		"checksum-check":     &ChecksumCheck{},
		"download-archive":   &DownloadArchive{},
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

func CreateGlacierClient() *glacier.Glacier {
	sess := createSession()
	return glacier.New(sess)
}

func CreateSqsClient() *sqs.SQS {
	sess := createSession()
	return sqs.New(sess)
}
