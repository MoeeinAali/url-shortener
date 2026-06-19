// Package redis is the read-side adapter. It implements the query read model and
// the projector's read-model writer on top of Redis.
package redis

import (
	"context"
	"errors"
	"strconv"

	goredis "github.com/redis/go-redis/v9"

	"url-shortener/internal/application/port"
	"url-shortener/internal/domain/link"
)

const (
	linkPrefix  = "link:"
	clickPrefix = "clicks:"
)

// ReadModel is the Redis-backed fast read store.
type ReadModel struct {
	rdb *goredis.Client
}

// Open connects to Redis.
func Open(addr, password string) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
	})
}

// NewReadModel constructs the adapter.
func NewReadModel(rdb *goredis.Client) *ReadModel {
	return &ReadModel{rdb: rdb}
}

// LongURL resolves a short code to its destination (query side).
func (r *ReadModel) LongURL(ctx context.Context, shortCode string) (string, error) {
	v, err := r.rdb.Get(ctx, linkPrefix+shortCode).Result()
	if errors.Is(err, goredis.Nil) {
		return "", link.ErrLinkNotFound
	}
	if err != nil {
		return "", err
	}
	return v, nil
}

// Clicks returns the click counter for an existing link (query side).
func (r *ReadModel) Clicks(ctx context.Context, shortCode string) (int64, error) {
	exists, err := r.rdb.Exists(ctx, linkPrefix+shortCode).Result()
	if err != nil {
		return 0, err
	}
	if exists == 0 {
		return 0, link.ErrLinkNotFound
	}

	n, err := r.rdb.Get(ctx, clickPrefix+shortCode).Int64()
	if errors.Is(err, goredis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return n, nil
}

// SaveLink upserts the short→long mapping (projector side).
func (r *ReadModel) SaveLink(ctx context.Context, shortCode, longURL string) error {
	return r.rdb.Set(ctx, linkPrefix+shortCode, longURL, 0).Err()
}

// DeleteLink removes the mapping so the link stops resolving (projector side).
func (r *ReadModel) DeleteLink(ctx context.Context, shortCode string) error {
	return r.rdb.Del(ctx, linkPrefix+shortCode).Err()
}

// SetClicks sets the counter to an authoritative value (projector side).
func (r *ReadModel) SetClicks(ctx context.Context, shortCode string, count int64) error {
	return r.rdb.Set(ctx, clickPrefix+shortCode, strconv.FormatInt(count, 10), 0).Err()
}

var (
	_ port.ReadModel       = (*ReadModel)(nil)
	_ port.ReadModelWriter = (*ReadModel)(nil)
)
