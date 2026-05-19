package interfacegeneric

import (
	"3za-digital/pkg/filter"
	"context"
)

type GenericRepository[T any] interface {
	Store(ctx context.Context, data T) error
	GetByID(ctx context.Context, id string) (T, error)
	GetAll(ctx context.Context, params filter.BaseParams) ([]T, int64, error)
	Update(ctx context.Context, data T) error
	Delete(ctx context.Context, id string) error
}
