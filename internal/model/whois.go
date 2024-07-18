package model

type WhoIs struct {
	ID       uint   `gorm:"primaryKey"`
	DomainID uint   `gorm:"index"`
	Content  string `gorm:"type:text;uniqueIndex"`
	ScanTime int64
}
