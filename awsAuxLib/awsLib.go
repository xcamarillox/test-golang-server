package awsAuxLib

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var S3 *S3Client

func init() {
	S3 = new(S3Client)
}

type S3Client struct {
	Region string
	Sess   *session.Session
	Svc    *s3.S3
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

// Crea una nueva sesión
func (t *S3Client) NewSession(region string) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	if err != nil {
		exitErrorf("PROBLEMA DE SESSION CON S3, %v", err)
	}
	t.Sess = sess
	t.Svc = s3.New(t.Sess)
}

/*
	Lista todo lo que se pueda ver en s3 con la cuenta de AWS
*/
func (t *S3Client) Ls() {
	result, err := t.Svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
}

/* Upload ó UploadObject:
Sube un archivo a s3

- filename string archivo local que deseas subir
- myBucket string nombre del bucket en tu cuenta de s3
- keyName string nombre del objeto final con la ruta completa pero sin el nombre del bucket
*/

func (t *S3Client) Upload(filename string, myBucket string, keyName string) {
	uploader := s3manager.NewUploader(t.Sess)
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to open file %q, %v", filename, err))
	}
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
		Body:   f,
	})
	if err != nil {
		fmt.Println(fmt.Errorf("failed to upload file, %v", err))
	}
	fmt.Println(result)
}

func (t *S3Client) UploadObject(filename string, myBucket string, keyName string) (resp *s3.PutObjectOutput) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	resp, err = t.Svc.PutObject(&s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
	})

	if err != nil {
		panic(err)
	}
	return resp
}

/* DeleteObject:
Borra un archivo del bucket de s3

- filename string archivo local que deseas subir
- myBucket string nombre del bucket en tu cuenta de s3
- keyName string nombre del objeto final con la ruta completa pero sin el nombre del bucket
*/

func (t *S3Client) DeleteObject(filename string, myBucket string, keyName string) (resp *s3.DeleteObjectOutput) {
	resp, err := t.Svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
	return resp
}

/*
	Genera URL publico del objeto que se encuentre en el bucket (myBucket)
	y que tenga el nombre que se ponga en (keyName)

	- myBucket string nombre del bucket en tu cuenta de s3
	- keyName string nombre del objeto a descargar con la ruta
				completa pero sin el nombre del bucket
*/
func (t *S3Client) GenerateUrl(myBucket string, keyName string) string {
	req, _ := t.Svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
	})
	urlStr, err := req.Presign(15 * time.Minute)

	if err != nil {
		log.Println("Failed to sign request", err)
	}
	//fmt.Println(urlStr)
	return urlStr

}
