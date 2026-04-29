package models

import "time"

// SegmentIdInfo 号段发号器配置表
type SegmentIdInfo struct {
	BizTag    string    `gorm:"column:biz_tag;primaryKey;type:varchar(32);not null;comment:业务标识"`
	MaxID     int64     `gorm:"column:max_id;type:bigint;not null;default:0;comment:当前已分配的最大ID"`
	Step      int       `gorm:"column:step;type:int;not null;default:1000;comment:每次分配的号段步长"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp;autoUpdateTime;comment:更新时间"`
}

// TableName 指定表名
func (SegmentIdInfo) TableName() string {
	return "segment_id_info"
}
