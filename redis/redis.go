package redis

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

type Client struct {
	*redis.Pool
}

func NewClient(url string) *Client {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(url)
		},
	}

	return &Client{pool}
}

func (c *Client) Call(ctx context.Context, fn func(redis.Conn) error) error {
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	return fn(conn)
}
