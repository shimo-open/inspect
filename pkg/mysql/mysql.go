package mysql

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Model struct {
	ID    int                   `json:"id" gorm:"not null;primary_key;auto_increment"`
	Ctime int64                 `gorm:"column:ctime;comment:创建时间" json:"ctime"`
	Utime int64                 `gorm:"column:utime;comment:更新时间" json:"utime"`
	Dtime soft_delete.DeletedAt `gorm:"column:dtime;comment:删除时间" json:"dtime"`
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	m.Ctime = time.Now().Unix()
	return nil
}

func (m *Model) BeforeSave(tx *gorm.DB) error {
	m.Utime = time.Now().Unix()
	return nil
}
