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
	Lista todos los buckets de AWS S3
*/
func (t *S3Client) ListBuckets() error {
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
Sube un archivo a s3 (las dos funciones son alternativas)

- filename string archivo local que deseas subir
- myBucket string nombre del bucket en tu cuenta de s3
- keyName string nombre del objeto final con la ruta completa pero sin el nombre del bucket
 Nota: La hacer un request sobre archivos con el ACL "ObjectCannedACLPublicRead" se debe agregar
 en la cabezera de la solicitud header = x-amz-acl, value = public-read
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
		ACL:    aws.String(s3.ObjectCannedACLPublicRead),
	})
	return resp, err
}

func (t *S3Client) UploadObject(filenameAndPath string, myBucket string, keyName string) (*s3.PutObjectOutput, error) {
	f, err := os.Open(filenameAndPath)
	if err != nil {
		return nil, err
	}
	resp, err := t.Svc.PutObject(&s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
		ACL:    aws.String(s3.ObjectCannedACLPublicRead),
	})
	return resp, err
}

/* DeleteObject:
Borra un archivo del bucket de s3

- myBucket string nombre del bucket en tu cuenta de s3
- keyName string nombre del objeto final con la ruta completa
*/

func (t *S3Client) DeleteObject(myBucket string, keyName string) (*s3.DeleteObjectOutput, error) {
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

func (t *S3Client) GetTemporalUrl(myBucket string, keyName string) (string, error) {
	req, _ := t.Svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
	})
	urlStr, err := req.Presign(30 * time.Minute)
	return urlStr, err
}

/* GetPresignedURL:
Obtiene un presignedURL

- myBucket string nombre del bucket en tu cuenta de s3
- keyName string nombre del objeto final con la ruta completa pero sin el nombre del bucket
Nota: La hacer un request sobre archivos con el ACL "ObjectCannedACLPublicRead" se debe agregar
 en la cabezera de la solicitud header = x-amz-acl, value = public-read
*/

func (t *S3Client) GetAPresignedURL(myBucket string, keyName string) (string, error) {
	req, _ := t.Svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(myBucket),
		Key:    aws.String(keyName),
		ACL:    aws.String(s3.ObjectCannedACLPublicRead),
	})
	urlStr, err := req.Presign(5 * time.Minute)
	return urlStr, err
}
