package crud

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
)

type (
	GORMModel interface {
		PrimaryKey() any
	}
	GenericCRUD[T GORMModel] struct {
		logger *log.Logger
		db     *gorm.DB
		omit   []string
	}
)

var (
	MultipleResultsError = errors.New("multiple results found")
	OmitByDefault        = []string{"CreatedAt", "UpdatedAt"}
)

func New[T GORMModel](db *gorm.DB, omit ...string) GenericCRUD[T] {
	return GenericCRUD[T]{
		logger: nil,
		db:     db,
		omit:   append(OmitByDefault, omit...),
	}
}

func (g GenericCRUD[T]) Create(ctx context.Context, v T, omit ...string) (*T, error) {
	err := g.db.Debug().WithContext(ctx).Omit(append(g.omit, omit...)...).Create(&v).Error
	return &v, err
}

func (g GenericCRUD[T]) GetOrCreate(ctx context.Context, v T, omit ...string) (*T, error) {
	err := g.db.Debug().WithContext(ctx).Omit(append(g.omit, omit...)...).Where(&v).FirstOrCreate(&v).Error
	return &v, err
}

func (g GenericCRUD[T]) GetByID(ctx context.Context, v T) (*T, error) {
	err := g.db.Debug().WithContext(ctx).Take(&v, v.PrimaryKey()).Error
	return &v, err
}

func (g GenericCRUD[T]) Query(ctx context.Context, v T, omit ...string) ([]*T, error) {
	var res []*T
	err := g.db.Debug().WithContext(ctx).Omit(append(g.omit, omit...)...).Where(&v).Find(&res).Error
	return res, err
}

func (g GenericCRUD[T]) QueryOne(ctx context.Context, v T, omit ...string) (*T, error) {
	var res []*T
	err := g.db.Debug().WithContext(ctx).Omit(append(g.omit, omit...)...).Where(&v).Find(&res).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("db error: %w", err)
		}
		return nil, err
	}
	if len(res) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if len(res) > 1 {
		return nil, MultipleResultsError
	}
	return res[0], nil
}

func (g GenericCRUD[T]) UpdateField(ctx context.Context, v T, column string, value any) error {
	return g.db.Debug().WithContext(ctx).Omit(g.omit...).Model(&v).Update(column, value).Error
}

func (g GenericCRUD[T]) Update(ctx context.Context, v T, omit ...string) (err error) {
	return g.db.Debug().WithContext(ctx).Omit(append(g.omit, omit...)...).Updates(&v).Error
}

func (g GenericCRUD[T]) UpdateMap(ctx context.Context, v T, m map[string]any) error {
	return g.db.Debug().WithContext(ctx).Model(&v).Updates(m).Error
}

func (g GenericCRUD[T]) Delete(ctx context.Context, v T) error {
	return g.db.Debug().WithContext(ctx).Delete(&v, v.PrimaryKey()).Error
}
