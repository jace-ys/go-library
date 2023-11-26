package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

type Client struct {
	db *sqlx.DB
}

func NewClient(ctx context.Context, connectionURL string) (*Client, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", connectionURL)
	if err != nil {
		return nil, err
	}

	db.MapperFunc(toSnakeCase)

	return &Client{db: db}, nil
}

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) Transact(ctx context.Context, fn func(*sqlx.Tx) error) (txErr error) {
	tx, err := c.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			txErr = c.rollback(tx, txErr)
			panic(p)
		} else if txErr != nil {
			txErr = c.rollback(tx, txErr)
		} else {
			txErr = tx.Commit()
		}
	}()

	return fn(tx)
}

func (c *Client) rollback(tx *sqlx.Tx, txErr error) error {
	err := tx.Rollback()
	if err != nil {
		return fmt.Errorf("tx err: %w, rollback err: %w", txErr, err)
	}
	return txErr
}
