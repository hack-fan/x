package xobj

import (
	"io"
	"net/http"

	"github.com/hack-fan/x/xerr"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// ProviderCOS provider tencent cloud cos
const (
	ProviderCOS = "cos"
)

var ErrorMissingKey = xerr.Newf(400, "MissingKey", "key is required")

// Config you can choose a provider and omit others
type Config struct {
	Provider string `default:"cos"`
	Cos      CosConfig
}

// Client is obj client interface
type Client interface {
	// Group can auto add prefix to key
	Group(prefix string) Client
	// Prefix show group prefix
	Prefix() string
	// GetRaw get raw http resp
	GetRaw(key string) (*http.Response, error)
	// GetReader please close it after use
	GetReader(key string) (io.ReadCloser, error)
	// Get get file byte content only
	Get(key string) ([]byte, error)
	// Exists exists or not
	Exists(key string) (bool, error)
	// Put please close reader yourself, key is required, default content type is application/octet-stream
	Put(r io.Reader, key, name, contentType string) error
	// PutRaw put the raw data without any meta, you can only use it by api.
	PutRaw(r io.Reader, key string) error
	// PutURL save the src online file to key
	PutURL(src, key string) error
	// Delete delete
	Delete(key string) error
	// IsNotFoundError check not found err
	IsNotFoundError(err error) bool
}

// New create a client
func New(config Config) Client {
	switch config.Provider {
	case ProviderCOS:
		return newCosClient(config.Cos)
	default:
		panic("invalid provider")
	}
}

func IsNotFoundError(err error) bool {
	// cos
	if cos.IsNotFoundError(err) {
		return true
	}
	// others

	// finally
	return false
}
