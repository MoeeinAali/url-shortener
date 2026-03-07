package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type ReadRepository struct {
	rdb *redis.Client
}

func NewReadRepository(r *redis.Client) *ReadRepository {
	return &ReadRepository{rdb: r}
}

func (r *ReadRepository) SaveLink(ctx context.Context, short string, long string) error {
	return r.rdb.Set(ctx, "link:"+short, long, 0).Err()
}

func (r *ReadRepository) DeleteLink(ctx context.Context, short string) error {
	return r.rdb.Del(ctx, "link:"+short).Err()
}

func (r *ReadRepository) GetLink(ctx context.Context, short string) (string, error) {
	return r.rdb.Get(ctx, "link:"+short).Result()
}

func (r *ReadRepository) IncClicks(ctx context.Context, short string) error {
	return r.rdb.Incr(ctx, "clicks:"+short).Err()
}

func (r *ReadRepository) GetClicks(ctx context.Context, short string) (int64, error) {
	return r.rdb.Get(ctx, "clicks:"+short).Int64()
}
