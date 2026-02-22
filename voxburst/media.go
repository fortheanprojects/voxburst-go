package voxburst

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// MediaService handles media-related API calls.
type MediaService struct {
	client *Client
}

// RequestUploadParams are the parameters for requesting an upload URL.
type RequestUploadParams struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

// ListMediaParams are the parameters for listing media.
type ListMediaParams struct {
	Status MediaStatus
	Limit  int
	Cursor string
}

// RequestUploadURL requests a presigned URL for uploading media.
func (s *MediaService) RequestUploadURL(ctx context.Context, params RequestUploadParams, opts ...RequestOption) (*UploadURL, error) {
	var result UploadURL
	if err := s.client.post(ctx, "/media/upload", params, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves media by ID.
func (s *MediaService) Get(ctx context.Context, id string, opts ...RequestOption) (*Media, error) {
	var media Media
	if err := s.client.get(ctx, "/media/"+id, &media, opts...); err != nil {
		return nil, err
	}
	return &media, nil
}

// List lists media with optional filtering.
func (s *MediaService) List(ctx context.Context, params *ListMediaParams, opts ...RequestOption) (*ListResponse[Media], error) {
	queryParams := make(map[string]string)
	if params != nil {
		if params.Status != "" {
			queryParams["status"] = string(params.Status)
		}
		if params.Limit > 0 {
			queryParams["limit"] = strconv.Itoa(params.Limit)
		}
		if params.Cursor != "" {
			queryParams["cursor"] = params.Cursor
		}
	}

	path := "/media" + buildQueryString(queryParams)
	var result ListResponse[Media]
	if err := s.client.get(ctx, path, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes media by ID.
func (s *MediaService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	return s.client.delete(ctx, "/media/"+id, nil, opts...)
}

// Upload is a convenience method that requests an upload URL and uploads the file.
// Returns the media ID that can be used in posts.
func (s *MediaService) Upload(ctx context.Context, filename, contentType string, data []byte, opts ...RequestOption) (string, error) {
	// Request upload URL
	uploadResp, err := s.RequestUploadURL(ctx, RequestUploadParams{
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   int64(len(data)),
	}, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}

	// Upload to S3
	if err := s.uploadToS3(ctx, uploadResp.UploadURL, uploadResp.UploadHeaders, data); err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return uploadResp.MediaID, nil
}

// UploadFile uploads a file from disk.
func (s *MediaService) UploadFile(ctx context.Context, filePath string, opts ...RequestOption) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	filename := filepath.Base(filePath)
	contentType := detectContentType(filename, data)

	// Request upload URL
	uploadResp, err := s.RequestUploadURL(ctx, RequestUploadParams{
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   stat.Size(),
	}, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}

	// Upload to S3
	if err := s.uploadToS3(ctx, uploadResp.UploadURL, uploadResp.UploadHeaders, data); err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return uploadResp.MediaID, nil
}

// UploadReader uploads from an io.Reader.
func (s *MediaService) UploadReader(ctx context.Context, filename, contentType string, size int64, reader io.Reader, opts ...RequestOption) (string, error) {
	// Request upload URL
	uploadResp, err := s.RequestUploadURL(ctx, RequestUploadParams{
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   size,
	}, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}

	// Read all data (needed for S3 upload)
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}

	// Upload to S3
	if err := s.uploadToS3(ctx, uploadResp.UploadURL, uploadResp.UploadHeaders, data); err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return uploadResp.MediaID, nil
}

// uploadToS3 uploads data to the presigned S3 URL.
func (s *MediaService) uploadToS3(ctx context.Context, uploadURL string, headers map[string]string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewReader(data))
	if err != nil {
		return err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("S3 upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// WaitForProcessing polls until the media is ready or fails.
func (s *MediaService) WaitForProcessing(ctx context.Context, mediaID string, opts ...RequestOption) (*Media, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		media, err := s.Get(ctx, mediaID, opts...)
		if err != nil {
			return nil, err
		}

		switch media.Status {
		case MediaStatusReady:
			return media, nil
		case MediaStatusFailed:
			return nil, &APIError{
				Code:    "MEDIA_PROCESSING_FAILED",
				Message: "Media processing failed",
			}
		}

		// Wait before polling again
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Poll every 2 seconds
			// Using default case to avoid blocking if context is cancelled
		}
	}
}

// detectContentType attempts to detect the content type from filename and data.
func detectContentType(filename string, data []byte) string {
	// Try by extension first
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	}

	// Fall back to http.DetectContentType
	return http.DetectContentType(data)
}
