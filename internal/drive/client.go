package drive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	gdrive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type DriveClient interface {
	UploadFile(ctx context.Context, path string) (*gdrive.File, error)
	GetFile(ctx context.Context, fileID string) (*gdrive.File, error)
}

type Client struct {
	svc *gdrive.Service
}

func NewClient(ctx context.Context, tokenSource oauth2.TokenSource) (*Client, error) {
	httpClient := oauth2.NewClient(ctx, tokenSource)
	svc, err := gdrive.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create drive service: %w", err)
	}
	return &Client{svc: svc}, nil
}

func (c *Client) UploadFile(ctx context.Context, path string) (*gdrive.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open upload file %s: %w", path, err)
	}
	defer f.Close()

	name := filepath.Base(path)
	file := &gdrive.File{Name: name}
	out, err := c.svc.Files.Create(file).Media(f).SupportsAllDrives(true).Fields("id,name,mimeType,webViewLink").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}
	return out, nil
}

func (c *Client) GetFile(ctx context.Context, fileID string) (*gdrive.File, error) {
	out, err := c.svc.Files.Get(fileID).SupportsAllDrives(true).Fields("id,name,mimeType,webViewLink").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get file %s: %w", fileID, err)
	}
	return out, nil
}
