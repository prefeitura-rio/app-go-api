package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// OrgaoSnapshot represents a cached snapshot of orgao data from RMI API
type OrgaoSnapshot struct {
	ID           int           `json:"id" gorm:"primaryKey;autoIncrement"`
	OrgaoID      string        `json:"orgao_id" gorm:"type:varchar(100);uniqueIndex;not null"`
	Name         string        `json:"name" gorm:"type:text;not null"`
	Sigla        *string       `json:"sigla,omitempty" gorm:"type:varchar(50)"`
	Metadata     OrgaoMetadata `json:"metadata,omitempty" gorm:"type:jsonb"`
	LastSyncedAt time.Time     `json:"last_synced_at" gorm:"not null;index:idx_orgao_snapshots_last_synced"`
	SyncStatus   string        `json:"sync_status" gorm:"type:varchar(50);not null;default:synced;index:idx_orgao_snapshots_status"`
	SyncError    *string       `json:"sync_error,omitempty" gorm:"type:text"`
	CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

// OrgaoMetadata stores flexible JSONB data from RMI API
type OrgaoMetadata map[string]interface{}

// TableName specifies the table name for GORM
func (OrgaoSnapshot) TableName() string {
	return "orgao_snapshots"
}

// Scan implements sql.Scanner interface for JSONB
func (m *OrgaoMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(OrgaoMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	result := make(OrgaoMetadata)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*m = result
	return nil
}

// Value implements driver.Valuer interface for JSONB
func (m OrgaoMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Sync status constants
const (
	SyncStatusSynced  = "synced"
	SyncStatusFailed  = "failed"
	SyncStatusPending = "pending"
)
