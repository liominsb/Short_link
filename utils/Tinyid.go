package utils

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

// IDGenerator 号段发号器
type IDGenerator struct {
	db     *gorm.DB
	bizTag string

	mutex   sync.Mutex // 保护内存号段并发安全
	current int64      // 本地内存当前发到的 ID
	max     int64      // 本地内存号段的上限 ID
}

// NewIDGenerator 初始化发号器
func NewIDGenerator(db *gorm.DB, bizTag string) *IDGenerator {
	return &IDGenerator{
		db:     db,
		bizTag: bizTag,
	}
}

// NextID 获取下一个全局唯一 ID
func (g *IDGenerator) NextID() (int64, error) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// 如果号段未初始化或已耗尽，向数据库申请新号段
	if g.current >= g.max {
		if err := g.loadNextSegment(); err != nil {
			return 0, err
		}
	}

	g.current++
	return g.current, nil
}

// loadNextSegment 利用 GORM 事务和悲观锁申请新号段
func (g *IDGenerator) loadNextSegment() error {
	return g.db.Transaction(func(tx *gorm.DB) error {
		var segment SegmentIdInfo

		// 1. 查询当前业务号段并加上排他锁
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("biz_tag = ?", g.bizTag).
			First(&segment).Error

		// 2. 核心修正：处理记录不存在的情况
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 如果记录不存在，则初始化一条记录
				segment = SegmentIdInfo{
					BizTag: g.bizTag,
					MaxID:  1000, // 初始 ID 从 0 开始（发出的第一个号将是 1）
					Step:   10,   // 默认步长
				}
				// 写入数据库
				if createErr := tx.Create(&segment).Error; createErr != nil {
					return fmt.Errorf("自动初始化号段记录失败: %w", createErr)
				}
				// 初始化成功后，不需要走后续的 Update 逻辑了，直接在内存分配即可
			} else {
				// 如果是其他数据库错误（如断网），直接返回
				return fmt.Errorf("查询并锁定号段记录失败: %w", err)
			}
		} else {
			// 3. 如果记录存在，执行正常的步长推进逻辑
			startID := g.current
			newMaxID := startID + int64(segment.Step)

			err = tx.Table("segment_id_info").
				Where("biz_tag = ?", g.bizTag).
				UpdateColumn("max_id", newMaxID).Error
			if err != nil {
				return fmt.Errorf("更新数据库 max_id 失败: %w", err)
			}
			segment.MaxID = startID // 为了下一步统一赋值给内存
		}

		// 4. 将新号段（或刚初始化的号段）加载到当前实例的内存中
		g.current = segment.MaxID
		g.max = segment.MaxID + int64(segment.Step)

		return nil
	})
}
