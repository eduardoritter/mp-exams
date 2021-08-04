package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
)

func main() {

	sess := ConnectAws()
	router := gin.Default()

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	router.Use(func(c *gin.Context) {
		c.Set("sess", sess)
		c.Next()
	})

	router.POST("/upload", UploadFiles)
	router.Run(":8080")
}

func ConnectAws() *session.Session {
	AccessKeyID := ""
	SecretAccessKey := ""
	MyRegion := "us-east-1"

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(MyRegion),
			Credentials: credentials.NewStaticCredentials(
				AccessKeyID,
				SecretAccessKey,
				"", // a token will be created when the session it's used.
			),
		})

	if err != nil {
		panic(err)
	}

	return sess
}

func UploadFiles(c *gin.Context) {

	sess := c.MustGet("sess").(*session.Session)
	uploader := s3manager.NewUploader(sess)

	bucket := ("")

	form, err := c.MultipartForm()
	if err != nil {
		return
	}

	files := form.File["files"]
	objects := []s3manager.BatchUploadObject{}

	for _, file := range files {

		src, err := file.Open()
		if err != nil {
			return
		}
		defer src.Close()

		var bUpload s3manager.BatchUploadObject

		bUpload.Object = &s3manager.UploadInput{
			Bucket: aws.String(bucket),
			ACL:    aws.String("private"),
			Key:    aws.String(file.Filename),
			Body:   src,
		}

		objects = append(objects, bUpload)
	}

	iter := &s3manager.UploadObjectsIterator{Objects: objects}

	if err := uploader.UploadWithIterator(aws.BackgroundContext(), iter); err != nil {
		return
	}

}
