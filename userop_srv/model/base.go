package model

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        int32          `gorm:"primary_key;AUTO_INCREMENT;not null" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
	IsDelete  bool           `gorm:"column:is_delete"`
}
type GormList []string

func (g *GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}
func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &g)
}
