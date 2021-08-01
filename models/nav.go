package models

import (
	"errors"
	"time"
)

type Navs struct {
	ID        uint      `gorm:"primarykey"`
	FundId    int       `json:"fund_id" gorm:"not null;index:idx_fund_id"`
	Fund      Funds     `json:"-" gorm:"association_autoupdate:false;references:FundId"`
	Date      time.Time `json:"-" gorm:"type:date;index:idx_date"`
	Timestamp int       `json:"timestamp" gorm:"type:int;not null;index:idx_timestamp"`
	Value     float64   `json:"value" gorm:"not null;precision:9;scale:4"`
}

func init() {
	DBAutoMigrate = append(DBAutoMigrate, &Navs{})
}

func navCreate(mdl *Navs) error {
	return DB.Create(mdl).Error
}

func navUpdate(mdl *Navs) error {
	return DB.Save(mdl).Error
}

func NavCreateOrUpdate(mdl *Navs) error {
	if mdl == nil || mdl.FundId <= 0 {
		return errors.New("invalid nav")
	}

	if err := DB.Where("fund_id=? AND timestamp=?", mdl.FundId, mdl.Timestamp).Find(mdl).Error; err == nil {
		//Update
		return navUpdate(mdl)
	} else {
		//Create
		return navCreate(mdl)
	}
}

func NavGetByFundId(fundId int) ([]Navs, error) {
	var navs []Navs
	result := DB.Where("fund_id=?", fundId).Find(&navs)

	return navs, result.Error
}
