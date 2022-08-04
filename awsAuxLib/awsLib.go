package awsAuxLib

import (
	"fmt"
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
func (t *S3Client) NewSession(region string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		exitErrorf("PROBLEMA DE SESSION CON S3, %v", err)
		return err
	}
	t.Sess = sess
	t.Svc = s3.New(t.Sess)
	return err
}

/*
	Lista todo lo que se pueda ver en s3 con la cuenta de AWS
*/
func (t *S3Client) Ls() error {
	result, err := t.Svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
		return err
	}
	fmt.Println("Buckets:")
	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}
	return err
}

/* Upload ó UploadObject:
Sube un archivo a s3

- filename string archivo local que deseas subir
- myBucket string nombre del bucket en tu cuenta de s3
- keyName string nombre del objeto final con la ruta completa pero sin el nombre del bucket
*/

func (t *S3Client) Upload(filename string, myBucket string, keyName string) (*s3manager.UploadOutput, error) {
	uploader := s3manager.NewUploader(t.Sess)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	resp, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
		Body:   f,
	})
	return resp, err
}

func (t *S3Client) UploadObject(filename string, myBucket string, keyName string) (*s3.PutObjectOutput, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	resp, err := t.Svc.PutObject(&s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
	})
	return resp, err
}

/* DeleteObject:
Borra un archivo del bucket de s3

- filename string archivo local que deseas subir
- myBucket string nombre del bucket en tu cuenta de s3
- keyName string nombre del objeto final con la ruta completa
*/

func (t *S3Client) DeleteObject(filename string, myBucket string, keyName string) (*s3.DeleteObjectOutput, error) {
	resp, err := t.Svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
	})
	return resp, err
}

/*  GetFileUrl:
Genera URL publico del objeto que se encuentre en el bucket (myBucket)
y que tenga el nombre que se ponga en (keyName)

- myBucket string nombre del bucket en tu cuenta de s3
- keyName string nombre del objeto a subur con la ruta completa
*/

func (t *S3Client) GetFileUrl(myBucket string, keyName string) (string, error) {
	req, _ := t.Svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
	})
	urlStr, err := req.Presign(30 * time.Minute)
	return urlStr, err
}

/* GetPresignedURL:
Obtiene un presignedURL

- filename string archivo local que deseas subir
- myBucket string nombre del bucket en tu cuenta de s3
- keyName string nombre del objeto final con la ruta completa pero sin el nombre del bucket
*/

func (t *S3Client) GetAPresignedURL(filename string, myBucket string, keyName string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	req, _ := t.Svc.PutObjectRequest(&s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
	})
	urlStr, err := req.Presign(5 * time.Minute)
	return urlStr, err
}
