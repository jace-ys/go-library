package postgres

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

type ClientConfig struct {
	ConnectionURL string
}

type Client struct {
	*sqlx.DB
}

func NewClient(url string) (*Client, error) {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("postgres connection failed: %w", err)
	}
	db.MapperFunc(toLowerSnakeCase)

	return &Client{
		DB: db,
	}, nil
}

func toLowerSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func (c *Client) Transact(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := c.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("postgres transaction failed: %w", err)
	}

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("postgres transaction failed: %w", err)
	}

	return tx.Commit()
}
