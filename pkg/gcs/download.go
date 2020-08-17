package gcs

import (
	"context"
	"io/ioutil"
	"time"

	"cloud.google.com/go/storage"
)

// CloudReadFileFromName slurps the contents of the named object into a byte array
func (c Client) CloudReadFileFromName(objectName string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, clientErr := storage.NewClient(ctx)
	if clientErr != nil {
		return nil, clientErr
	}

	reader, err := client.Bucket(c.BucketName).Object(objectName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return data, nil
}
