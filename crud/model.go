package crud

import (
	"gorm.io/gorm"
	"time"
)

/*
Model is base type like gorm.Model but with uint32 ID type for better compatibility with protobuf
Model implements PrimaryKey interface
Example:

	type User struct {
		Model
		Name string
	}
*/
type Model struct {
	ID        uint32 `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m Model) PrimaryKey() any {
	return m.ID
}
