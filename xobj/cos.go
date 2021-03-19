package xobj

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/hack-fan/x/xerr"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type CosConfig struct {
	AppID     string
	Region    string
	SecretID  string
	SecretKey string
	Bucket    string
}

type cosClient struct {
	c *cos.Client
	// just use bg context for cos api
	ctx context.Context
}

func newCosClient(config CosConfig) *cosClient {
	u, _ := url.Parse(fmt.Sprintf("http://%s-%s.cos.%s.myqcloud.com",
		config.Bucket, config.AppID, config.Region))
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretID,
			SecretKey: config.SecretKey,
		},
	})
	return &cosClient{
		c:   c,
		ctx: context.Background(),
	}
}

func (c *cosClient) GetRaw(key string) (*http.Response, error) {
	if key == "" {
		return nil, ErrorMissingKey
	}
	resp, err := c.c.Object.Get(c.ctx, key, nil)
	if err != nil {
		return nil, err
	}
	return resp.Response, nil
}

func (c *cosClient) GetReader(key string) (io.ReadCloser, error) {
	resp, err := c.GetRaw(key)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *cosClient) Get(key string) ([]byte, error) {
	reader, err := c.GetReader(key)
	if err != nil {
		return nil, err
	}
	file, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	reader.Close()
	return file, nil
}

// PutRaw put the raw data without any meta, you can only use it by api.
func (c *cosClient) PutRaw(r io.Reader, key string) error {
	return c.Put(r, key, "", "")
}

// Put a file with it's content type and download name from reader
func (c *cosClient) Put(r io.Reader, key, name, contentType string) error {
	if key == "" {
		return ErrorMissingKey
	}
	if name == "" {
		name = key
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentDisposition: fmt.Sprintf(`attachment; filename="%s"`, name),
			ContentType:        contentType,
		},
	}
	_, err := c.c.Object.Put(c.ctx, key, r, opt)
	if err != nil {
		return err
	}
	return nil
}

func (c *cosClient) Delete(key string) error {
	if key == "" {
		return xerr.New(400, "EmptyKey", "empty key")
	}
	_, err := c.c.Object.Delete(c.ctx, key)
	if err != nil {
		return err
	}
	return nil
}

// Exists check if the key exists in cos
func (c *cosClient) Exists(key string) (bool, error) {
	if key == "" {
		return false, ErrorMissingKey
	}
	_, err := c.c.Object.Head(c.ctx, key, nil)
	if cos.IsNotFoundError(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *cosClient) IsNotFoundError(err error) bool {
	return cos.IsNotFoundError(err)
}
