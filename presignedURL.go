package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	req, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}

	return req.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil || strings.TrimSpace(*video.VideoURL) == "" {
		return video, nil
	}

	parts := strings.SplitN(*video.VideoURL, ",", 2)
	if len(parts) != 2 {
		return database.Video{}, fmt.Errorf("invalid video_url format: %q", *video.VideoURL)
	}

	bucket := strings.TrimSpace(parts[0])
	key := strings.TrimSpace(parts[1])
	if bucket == "" || key == "" {
		return database.Video{}, fmt.Errorf("bucket or key is empty: %q", *video.VideoURL)
	}

	presignURL, err := generatePresignedURL(cfg.s3Client, bucket, key, 5*time.Minute)
	if err != nil {
		return database.Video{}, err
	}

	video.VideoURL = &presignURL
	return video, nil
}
