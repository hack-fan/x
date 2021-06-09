package xobj

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hack-fan/x/xerr"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type CosConfig struct {
	AppID     string
	Region    string
	SecretID  string
	SecretKey string
	Bucket    string
	Prefix    string
}

type cosClient struct {
	client *cos.Client
	httpc  *http.Client
	prefix string
	// just use bg context for cos api
	ctx context.Context
}

func newCosClient(config CosConfig) *cosClient {
	u, _ := url.Parse(fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com",
		config.Bucket, config.AppID, config.Region))
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretID,
			SecretKey: config.SecretKey,
		},
		Timeout: time.Second * 30,
	})
	return &cosClient{
		client: c,
		httpc:  &http.Client{Timeout: time.Second * 30},
		prefix: config.Prefix,
		ctx:    context.Background(),
	}
}

func (c *cosClient) Group(prefix string) Client {
	return &cosClient{
		client: c.client,
		httpc:  c.httpc,
		prefix: c.prefix + prefix,
		ctx:    context.Background(),
	}
}

func (c *cosClient) Prefix() string {
	return c.prefix
}

func (c *cosClient) GetRaw(key string) (*http.Response, error) {
	if key == "" {
		return nil, ErrorMissingKey
	}
	resp, err := c.client.Object.Get(c.ctx, c.prefix+key, nil)
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

func (c *cosClient) PutURL(src, key string) error {
	resp, err := c.httpc.Get(src)
	if err != nil {
		return fmt.Errorf("get %s failed when put to cos: %w", src, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get %s failed when put to cos, status: %s, error: %s", src, resp.Status, string(msg))
	}
	tp := resp.Header.Get("Content-Type")
	name := resp.Header.Get("Content-Disposition")
	t, _, err := mime.ParseMediaType(tp)
	if err != nil {
		return fmt.Errorf("parse %s media type failed when put to cos: %w", src, err)
	}
	if strings.HasPrefix(t, "text") {
		return fmt.Errorf("target %s is text, failed to put to cos", src)
	}
	return c.Put(resp.Body, key, name, tp)
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
	_, err := c.client.Object.Put(c.ctx, c.prefix+key, r, opt)
	if err != nil {
		return err
	}
	return nil
}

func (c *cosClient) Delete(key string) error {
	if key == "" {
		return xerr.New(400, "EmptyKey", "empty key")
	}
	_, err := c.client.Object.Delete(c.ctx, c.prefix+key)
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
	_, err := c.client.Object.Head(c.ctx, c.prefix+key, nil)
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
