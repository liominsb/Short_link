package utils

import (
	"Short_link/models"
	"errors"
	"fmt"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IDGenerator 号段发号器
type IDGenerator struct {
	db     *gorm.DB
	bizTag string

	mutex   sync.Mutex // 保护内存号段并发安全
	current uint64     // 本地内存当前发到的 ID
	max     uint64     // 本地内存号段的上限 ID
}

// NewIDGenerator 初始化发号器
func NewIDGenerator(db *gorm.DB, bizTag string) *IDGenerator {
	return &IDGenerator{
		db:     db,
		bizTag: bizTag,
	}
}

// NextID 获取下一个全局唯一 ID
func (g *IDGenerator) NextID() (uint64, error) {
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
		var segment models.SegmentIdInfo

		// 1. 查询当前业务号段并加上排他锁
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("biz_tag = ?", g.bizTag).
			First(&segment).Error

		// 2. 分支 A：记录不存在，执行初始化
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				startID := uint64(1000)
				step := uint64(1000)
				newMaxID := startID + step

				// 初始化一条记录（直接存入占据后的上限）
				segment = models.SegmentIdInfo{
					BizTag: g.bizTag,
					MaxID:  int64(newMaxID),
					Step:   int(step),
				}
				if createErr := tx.Create(&segment).Error; createErr != nil {
					return fmt.Errorf("自动初始化号段记录失败: %w", createErr)
				}

				// 直接为内存赋值并返回
				g.current = startID
				g.max = newMaxID
				return nil
			}
			return fmt.Errorf("查询并锁定号段记录失败: %w", err)
		}

		// 3. 分支 B：记录存在，执行步长推进
		// 【绝对真理】：起点必须用数据库里查出的 segment.MaxID
		startID := uint64(segment.MaxID)
		newMaxID := startID + uint64(segment.Step)

		res := tx.Table("segment_id_info").
			Where("biz_tag = ?", g.bizTag).
			UpdateColumn("max_id", newMaxID)

		if res.Error != nil {
			return fmt.Errorf("更新数据库 max_id 失败: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("严重异常: 更新了 0 行数据，防止发号倒流")
		}

		// 4. 将新号段直接加载到内存中，没有任何多余的中间变量
		g.current = startID
		g.max = newMaxID

		return nil
	})
}
