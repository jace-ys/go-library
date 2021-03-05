package postgres

import (
	"context"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

type Client struct {
	*sqlx.DB
}

func NewClient(url string) (*Client, error) {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	db.MapperFunc(toLowerSnakeCase)
	return &Client{db}, nil
}

func toLowerSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func (c *Client) Transact(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := c.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
