package minio

import (
	"context"
	"io"

	minioClient "github.com/minio/minio-go/v7"
)

type BuyerAvatarRepoMinio struct {
	client *minioClient.Client
	bucket string
}

func NewBuyerAvatarRepo(client *minioClient.Client, bucket string) *BuyerAvatarRepoMinio {
	return &BuyerAvatarRepoMinio{client: client, bucket: bucket}
}

func (repo *BuyerAvatarRepoMinio) Put(ctx context.Context, objectName string, body io.Reader, size int64, contentType string) error {
	_, err := repo.client.PutObject(ctx, repo.bucket, objectName, body, size, minioClient.PutObjectOptions{ContentType: contentType})
	return err
}

func (repo *BuyerAvatarRepoMinio) Bucket() string {
	return repo.bucket
}
