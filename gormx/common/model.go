package common

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// BaseTimeModel 表示使用数据库时间类型的基础模型。
type BaseTimeModel struct {
	ID        int64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement;comment:主键ID"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime(3);not null;autoCreateTime:milli;comment:创建时间"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:datetime(3);not null;autoUpdateTime:milli;comment:更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index;comment:软删除时间，NULL表示未删除"`
}

// BaseInt64Model 表示使用 UTC 毫秒时间戳的基础模型。
type BaseInt64Model struct {
	ID        int64                 `gorm:"column:id;type:bigint;primaryKey;autoIncrement;comment:主键ID"`
	CreatedAt int64                 `gorm:"column:created_at;type:bigint;not null;autoCreateTime:milli;comment:创建时间，UTC毫秒时间戳"`
	UpdatedAt int64                 `gorm:"column:updated_at;type:bigint;not null;autoUpdateTime:milli;comment:更新时间，UTC毫秒时间戳"`
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;index;comment:软删除时间，UTC毫秒时间戳，0表示未删除"`
}
