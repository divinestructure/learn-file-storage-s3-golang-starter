package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType, aspectRatio string) string {
	base := make([]byte, 32)
	_, err := rand.Read(base)
	if err != nil {
		panic("failed to generate random bytes")
	}
	id := base64.RawURLEncoding.EncodeToString(base)

	ext := mediaTypeToExt(mediaType)
	if aspectRatio == "" {
		return fmt.Sprintf("%s%s", id, ext)
	}
	if aspectRatio == "16:9" {
		return fmt.Sprintf("landscape-%s%s", id, ext)
	}
	if aspectRatio == "9:16" {
		return fmt.Sprintf("portrait-%s%s", id, ext)
	}
	return fmt.Sprintf("other-%s%s", id, ext)

}

func (cfg apiConfig) getObjectURL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func getVideoAspectRatio(filePath string) (aspectRatio string, err error) {
	osCmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	var buffer bytes.Buffer
	osCmd.Stdout = &buffer

	if err := osCmd.Run(); err != nil {
		return "", err
	}

	var osResponse struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(buffer.Bytes(), &osResponse); err != nil {
		return "", err
	}

	for _, v := range osResponse.Streams {
		if v.CodecType == "video" {
			if v.Height == 0 || v.Width == 0 {
				return "", errors.New("invalid video ratio")
			}

			if v.Height*16 == v.Width*9 {
				return "16:9", nil
			}
			if v.Height*9 == v.Width*16 {
				return "9:16", nil
			}
			return "other", nil

		}

	}

	return "", errors.New("no video stream")
}
