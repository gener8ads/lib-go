package gcs

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
)

// Client is our own little client, containing the variables needed to interact with a bucket
type Client struct {
	BucketName string
}

// CloudUpload stores the file in a bucket
func (c Client) CloudUpload(file *multipart.FileHeader) (string, int64, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, clientErr := storage.NewClient(ctx)
	if clientErr != nil {
		return "", 0, "", clientErr
	}

	fileExt, err := mime.ExtensionsByType(file.Header["Content-Type"][0])
	if err != nil || len(fileExt) == 0 {
		return "", 0, "", err
	}
	objectName := fmt.Sprintf("%s%s", c.GenerateFolderName(), fileExt[0])
	writer := client.Bucket(c.BucketName).Object(objectName).NewWriter(ctx)
	defer writer.Close()

	reader, err := file.Open()
	if err != nil {
		return "", 0, "", err
	}

	_, uploadError := io.Copy(writer, reader)
	if uploadError != nil {
		return "", 0, "", uploadError
	}

	hasher := sha1.New()
	_, hashError := io.Copy(hasher, reader)
	if hashError != nil {
		return "", 0, "", hashError
	}
	fileHash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return fmt.Sprintf("gs://%s/%s", c.BucketName, objectName), file.Size, fileHash, nil
}

// CloudWriteFileWithName stores the file in a bucket, under a user provided name, which may include a folder-path
func (c Client) CloudWriteFileWithName(fileContents []byte, objectName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, clientErr := storage.NewClient(ctx)
	if clientErr != nil {
		return "", clientErr
	}

	writer := client.Bucket(c.BucketName).Object(objectName).NewWriter(ctx)
	defer writer.Close()
	reader := bytes.NewReader(fileContents)

	_, uploadError := io.Copy(writer, reader)
	if uploadError != nil {
		return "", uploadError
	}

	return fmt.Sprintf("gs://%s/%s", c.BucketName, objectName), nil
}

// GenerateFolderName to try and keep the number of objects in each folder down
func (c Client) GenerateFolderName() (path string) {
	id := uuid.New().String()
	strLen := len(id)
	path1 := id[strLen-1:]
	path2 := id[strLen-3 : strLen-1]
	path = fmt.Sprintf("%s/%s/%s", path1, path2, id)
	return path
}

// ObjectExists checks to see if the provided URI does in fact represent an object in the expeted GCS bucket
func (c Client) ObjectExists(objectURL string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, clientErr := storage.NewClient(ctx)
	if clientErr != nil {
		return false, clientErr
	}
	parts := strings.Split(objectURL, "/")
	imageName := strings.Join(parts[3:], "/")

	reader, err := client.Bucket(c.BucketName).Object(imageName).NewReader(ctx)
	if err != nil {
		return false, err
	}
	reader.Close()

	return true, nil
}

// DeleteObject removes an object from the GCS bucket
func (c Client) DeleteObject(objectURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, clientErr := storage.NewClient(ctx)
	if clientErr != nil {
		return clientErr
	}
	parts := strings.Split(objectURL, "/")
	imageName := strings.Join(parts[3:], "/")

	object := client.Bucket(c.BucketName).Object(imageName)
	if err := object.Delete(ctx); err != nil {
		return err
	}

	return nil
}
