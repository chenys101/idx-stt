package model

import "time"

type StockBase struct {
	Name          string  `json:"name"`
	ChangePercent float64 `json:"change_percent"`
	ChangePrice   float64 `json:"change_price"`
}
type StockBaseEn struct {
	Name          string  `json:"name"`
        Cp string `json:"cp,omitempty"`
        Cpr   string `json:"cpr,omitempty"`
}
type StockMonitor struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	Code         string    `gorm:"uniqueIndex"`
	MonitorValue string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
