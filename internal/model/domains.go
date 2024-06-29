package model

type Domain struct {
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	TLD  string `gorm:"Index"`
	Name string `gorm:"uniqueIndex"`
}
