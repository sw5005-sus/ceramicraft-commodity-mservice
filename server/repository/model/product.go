package model

import "gorm.io/gorm"

type Product struct {
	gorm.Model

	Name             string `gorm:"type:varchar(255);not null"`
	Category         string `gorm:"type:varchar(255);not null"`
	Price            int64  `gorm:"type:int;not null"`
	Desc             string `gorm:"type:text;not null"`
	Stock            int64  `gorm:"type:int;not null"`
	PicInfo          string `gorm:"type:text;not null"`
	Dimensions       string `gorm:"type:varchar(255)"`
	Material         string `gorm:"type:varchar(255)"`
	Weight           string `gorm:"type:varchar(255)"`
	Capacity         string `gorm:"type:varchar(255)"`
	CareInstructions string `gorm:"type:text"`
	Status           int32  `gorm:"type:int;not null"`           // 0: 未上架, 1: 已上架
	Version          int64  `gorm:"type:int;not null;default:0"` // 用于乐观锁
	LatestEditorId   int    `gorm:"type:int;not null"`           // 最新编辑者ID
	LatestReviewerId int    `gorm:"type:int;not null"`           // 审核者ID
}

func (Product) TableName() string {
	return "products"
}
