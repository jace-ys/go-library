package redis

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

type ClientConfig struct {
	ConnectionURL string
}

type Client struct {
	*redis.Pool
}

func NewClient(url string) (*Client, error) {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(url)
		},
	}

	return &Client{
		Pool: pool,
	}, nil
}

func (c *Client) Transact(ctx context.Context, fn func(redis.Conn) error) error {
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("redis transaction failed: %w", err)
	}
	defer conn.Close()

	if err := fn(conn); err != nil {
		return fmt.Errorf("redis transaction failed: %w", err)
	}

	return nil
}
