package models

import (
	"errors"
	"gorm.io/gorm"
)

type Funds struct {
	gorm.Model
	FundId      int     `json:"fund_id" gorm:"not null;index:idx_fund_id,unique"`
	FundName    string  `json:"fund_name" gorm:"not null"`
	FundType    string  `json:"fund_type" gorm:"index:idx_fund_type"`
	MiName      string  `json:"mi_name"`
	MiCode      string  `json:"mi_code" gorm:"index:idx_mi_code"`
	Mi          Manager `json:"mi" gorm:"association_autoupdate:false;foreignKey:MiCode"`
	LastNAV     float64 `json:"last_nav"`
	D1          float64 `json:"d1"`
	D3          float64 `json:"d3"`
	M1          float64 `json:"m1"`
	M3          float64 `json:"m3"`
	M6          float64 `json:"m6"`
	M9          float64 `json:"m9"`
	YTD         float64 `json:"ytd"`
	Y1          float64 `json:"y1"`
	Y3          float64 `json:"y3"`
	Y5          float64 `json:"y5"`
	HiLo        float64 `json:"hi_lo"`
	Sharpe      float64 `json:"sharpe"`
	DrawDown    float64 `json:"draw_down"`
	DdPeriode   int     `json:"dd_periode"`
	HistRisk    float64 `json:"hist_risk"`
	AUM         float64 `json:"aum"`
	Morningstar float64 `json:"morningstar"`
	Active      int     `json:"active" gorm:"index:idx_active"`
	Risk        string  `json:"risk"`
	Type        string  `json:"type"`
}

func init() {
	DBAutoMigrate = append(DBAutoMigrate, &Funds{})
}

func FundGetById(fundId int) (*Funds, error) {
	var m Funds
	result := DB.Where("id=?", fundId).Find(&m)

	if result.Error == nil && result.RowsAffected != 0 {
		return &m, nil
	} else {
		return nil, result.Error
	}
}

func fundCreate(mdl *Funds) error {
	return DB.Create(mdl).Error
}

func fundUpdate(mdl *Funds) error {
	return DB.Save(mdl).Error
}

func FundCreateOrUpdate(mdl *Funds) error {
	if mdl == nil || mdl.FundId <= 0 {
		return errors.New("invalid fund")
	}

	if err := DB.Where("fund_id=?", mdl.FundId).Find(mdl).Error; err == nil {
		//Update
		return fundUpdate(mdl)
	} else {
		//Create
		return fundCreate(mdl)
	}
}
