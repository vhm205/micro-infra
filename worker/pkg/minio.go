// Package minio provides functions for interacting with Minio
package minio

import (
	"context"
	"fmt"
	log "worker-service/utils"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	Client     *minio.Client
}

func (m *MinioClient) Connect() error {
	// Initialize minio client object.
	minioClient, err := minio.New(m.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(m.AccessKey, m.SecretKey, ""),
		Secure: false,
	})
	log.FailOnError(err, "Failed to connect to Minio")

	m.Client = minioClient
	fmt.Println("✅ Connected to Minio")

	return err
}

func (m *MinioClient) GetObject(key string) *minio.Object {
	ctx := context.Background()
	opts := minio.GetObjectOptions{}
	obj, err := m.Client.GetObject(ctx, m.BucketName, key, opts)
	log.FailOnError(err, "Failed to get object")
	return obj
}
