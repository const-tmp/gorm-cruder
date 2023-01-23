package crud

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
)

type (
	// GORMModel is an interface for getting primary keys
	GORMModel interface {
		PrimaryKey() any
	}

	// GenericCRUD is generic struct for model's CRUD operations
	GenericCRUD[T GORMModel] struct {
		logger *log.Logger
		db     *gorm.DB
		omit   []string
	}

	OrderBy uint
	Between struct {
		From, To any
	}

	Query struct {
		Omit    []string
		Preload []string
		OrderBy map[string]OrderBy
		Equal   map[string]string
		Like    map[string]string
		Between map[string]Between
	}
)

const (
	ASC OrderBy = 1 << iota
	DESC
)

func (ob OrderBy) String() string {
	if ob == ASC {
		return "ASC"
	}
	if ob == DESC {
		return "DESC"
	}
	panic("unknown OrderBy")
}

var (
	// MultipleResultsError is returned when GenericCRUD.QueryOne finds more than 1 row
	MultipleResultsError = errors.New("multiple results found")
)

// New is a constructor
func New[T GORMModel](db *gorm.DB, omit ...string) GenericCRUD[T] {
	return GenericCRUD[T]{
		logger: nil,
		db:     db,
		omit:   omit,
	}
}

// Create Model
func (g GenericCRUD[T]) Create(ctx context.Context, v T, omit ...string) (*T, error) {
	err := g.db.Debug().WithContext(ctx).Omit(append(g.omit, omit...)...).Create(&v).Error
	return &v, err
}

// GetOrCreate Model
func (g GenericCRUD[T]) GetOrCreate(ctx context.Context, v T, omit ...string) (*T, error) {
	err := g.db.Debug().WithContext(ctx).Omit(append(g.omit, omit...)...).Where(&v).FirstOrCreate(&v).Error
	return &v, err
}

// GetByID get Model by primary key; v MUST have non-zero primary key
func (g GenericCRUD[T]) GetByID(ctx context.Context, v T) (*T, error) {
	err := g.db.Debug().WithContext(ctx).Take(&v, v.PrimaryKey()).Error
	return &v, err
}

// Query by non-zero fields of v; returns slice of Model's
func (g GenericCRUD[T]) Query(ctx context.Context, v T, omit ...string) ([]*T, error) {
	var res []*T
	err := g.db.Debug().WithContext(ctx).Omit(append(g.omit, omit...)...).Where(&v).Find(&res).Error
	return res, err
}

// QueryOne by non-zero fields of v; returns exactly one Model or error
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

// QueryMap by non-zero fields of v; returns slice of Model's
func (g GenericCRUD[T]) QueryMap(ctx context.Context, q map[string]any, omit ...string) ([]*T, error) {
	var res []*T
	err := g.db.Debug().WithContext(ctx).Omit(omit...).Find(&res, q).Error
	return res, err
}

// QueryMapOne by non-zero fields of v; returns exactly one Model or error
func (g GenericCRUD[T]) QueryMapOne(ctx context.Context, q map[string]any, omit ...string) (*T, error) {
	res, err := g.QueryMap(ctx, q, omit...)
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

// SmartQuery by non-zero fields of v; returns slice of Model's
func (g GenericCRUD[T]) SmartQuery(ctx context.Context, q Query) ([]*T, error) {
	var (
		res  []*T
		err  error
		stmt = g.db.Debug().WithContext(ctx).Omit(q.Omit...)
	)
	for _, s := range q.Preload {
		stmt = stmt.Preload(s)
	}
	for k, v := range q.OrderBy {
		stmt = stmt.Order(k + " " + v.String())
	}
	for k, v := range q.Like {
		stmt = stmt.Where(k+" LIKE ?", fmt.Sprintf("%%%s%%", v))
	}
	for k, v := range q.Between {
		stmt = stmt.Where(k+" BETWEEN ? AND ?", v.From, v.To)
	}
	for k, v := range q.Equal {
		stmt = stmt.Where(k+" = ?", v)
	}
	err = stmt.Find(&res).Error
	return res, err
}

// SmartQueryOne by non-zero fields of v; returns exactly one Model or error
func (g GenericCRUD[T]) SmartQueryOne(ctx context.Context, q Query) (*T, error) {
	res, err := g.SmartQuery(ctx, q)
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

// UpdateField of Model; if v has non-zero primary key - filter by primary key
func (g GenericCRUD[T]) UpdateField(ctx context.Context, v T, column string, value any) error {
	return g.db.Debug().WithContext(ctx).Omit(g.omit...).Model(&v).Update(column, value).Error
}

// Update if v has non-zero primary key - filter by primary key
func (g GenericCRUD[T]) Update(ctx context.Context, v T, omit ...string) (err error) {
	return g.db.Debug().WithContext(ctx).Omit(append(g.omit, omit...)...).Updates(&v).Error
}

// UpdateMap if v has non-zero primary key - filter by primary key
func (g GenericCRUD[T]) UpdateMap(ctx context.Context, v T, q map[string]any) error {
	return g.db.Debug().WithContext(ctx).Model(&v).Updates(q).Error
}

// Delete if v has non-zero primary key - filter by primary key
func (g GenericCRUD[T]) Delete(ctx context.Context, v T) error {
	return g.db.Debug().WithContext(ctx).Delete(&v, v.PrimaryKey()).Error
}
