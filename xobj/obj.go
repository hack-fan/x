package xobj

import (
	"io"
	"net/http"

	"github.com/hack-fan/x/xerr"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// all providers
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
	// get raw http resp
	GetRaw(key string) (*http.Response, error)
	// please close it after use
	GetReader(key string) (io.ReadCloser, error)
	// get file byte content only
	Get(key string) ([]byte, error)
	// exists or not
	Exists(key string) (bool, error)
	// please close reader yourself, key is required, default content type is application/octet-stream
	Put(r io.Reader, key, name, contentType string) error
	PutRaw(r io.Reader, key string) error
	// delete
	Delete(key string) error
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
