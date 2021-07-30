package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

//Manager Investasi
type Manager struct {
	MiCode    string `json:"mi_code" gorm:"primarykey;autoIncrement:false"`
	MiName    string `json:"mi_name" gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func init() {
	DBAutoMigrate = append(DBAutoMigrate, &Manager{})
}

func ManagerCreateIfNotExists(m *Manager) (err error) {
	if m.MiCode == "" || m.MiName == "" {
		return errors.New("empty MI Code/Name")
	}

	if err = DB.First(m, "mi_code = ?", m.MiCode).Error; err == nil {
		// exists
		return
	} else {
		// Create
		return DB.Create(m).Error
	}
}
