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
		ID() any
	}
	GenericCRUD[T GORMModel] struct {
		logger *log.Logger
		db     *gorm.DB
	}
)

var (
	MultipleResultsError = errors.New("multiple results found")
)

func New[T GORMModel](db *gorm.DB) GenericCRUD[T] {
	return GenericCRUD[T]{
		logger: nil,
		db:     db,
	}
}

func (g GenericCRUD[T]) Create(ctx context.Context, v T) (newV T, err error) {
	newV = v
	err = g.db.Debug().WithContext(ctx).Create(&newV).Error
	return
}

func (g GenericCRUD[T]) GetOrCreate(ctx context.Context, v T) (newV T, err error) {
	err = g.db.Debug().WithContext(ctx).Where(&v).FirstOrCreate(&newV).Error
	return
}

func (g GenericCRUD[T]) GetByID(ctx context.Context, v T) (newV T, err error) {
	err = g.db.Debug().WithContext(ctx).Take(&newV, v.ID()).Error
	return
}

func (g GenericCRUD[T]) Query(ctx context.Context, v T) (newV []T, err error) {
	err = g.db.Debug().WithContext(ctx).Where(&v).Find(&newV).Error
	return
}

func (g GenericCRUD[T]) QueryOne(ctx context.Context, v T) (newV T, err error) {
	var res []T
	err = g.db.Debug().WithContext(ctx).Where(&v).Find(&res).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("db error: %w", err)
		}
		return *new(T), err
	}
	if len(res) == 0 {
		return *new(T), gorm.ErrRecordNotFound
	}
	if len(res) > 1 {
		return *new(T), MultipleResultsError
	}
	return res[0], nil
}

func (g GenericCRUD[T]) UpdateField(ctx context.Context, v T, column string, value any) (err error) {
	err = g.db.Debug().WithContext(ctx).Model(&v).Update(column, value).Error
	return
}

func (g GenericCRUD[T]) Update(ctx context.Context, v T) (err error) {
	err = g.db.Debug().WithContext(ctx).Updates(&v).Error
	return
}

func (g GenericCRUD[T]) UpdateMap(ctx context.Context, v map[string]any) (err error) {
	err = g.db.Debug().WithContext(ctx).Updates(&v).Error
	return
}

func (g GenericCRUD[T]) Delete(ctx context.Context, v map[string]any) (err error) {
	err = g.db.Debug().WithContext(ctx).Delete(&v).Error
	return
}
