package model

import "time"

// User represents a user in the database
// Note the use of json tags to control the output format for API responses.
// This allows the frontend to consistently use lowercase keys (e.g., "id", "name"),
// while Go code uses uppercase, exported fields (e.g., "ID", "Name").
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:text" json:"name"`
	Email     string    `gorm:"uniqueIndex" json:"email"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
}
