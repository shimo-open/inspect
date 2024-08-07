package mysql

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReportMeasure 衡量指标
type ReportMeasure struct {
	ID         int64   `gorm:"auto;not null;comment:主键ID" json:"id" form:"id"`                                                          // 主键ID
	Metric     string  `gorm:"size:64;not null;comment:指标ID;uniqueIndex:idx_target_metric_time" json:"metric" form:"metric"`            // 指标ID
	TargetName string  `gorm:"size:128;not null;comment:关联目标名称;uniqueIndex:idx_target_metric_time" json:"targetName" form:"targetName"` // 关联目标UUID，对于服务，可以是服务名
	TargetType string  `gorm:"size:32;not null;comment:关联目标类型;uniqueIndex:idx_target_metric_time" json:"targetType" form:"targetType"`  // 关联目标UUID，对于服务，可以是服务名
	Val        float64 `gorm:"size:12;not null;comment:指标值" json:"val" form:"val"`                                                      // 值
	Time       int64   `gorm:"type:bigint;not null;comment:时间戳;uniqueIndex:idx_target_metric_time" json:"time" form:"time"`             // 时间戳
	Ctime      int64   `gorm:"type:bigint;not null;comment:创建时间;autoCreateTime" json:"ctime" form:"ctime"`                              // 创建时间
}

func (ReportMeasure) TableName() string {
	return "report_measure"
}

// ReportMeasureCreateMulti 创建一条记录
func ReportMeasureCreateMulti(db *gorm.DB, data []ReportMeasure) (err error) {
	if err = db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(data, len(data)).Error; err != nil {
		err = fmt.Errorf("measure create err: %w", err)
		return
	}
	// 同时插入 target 表，记录唯一的 target_name target_type
	var (
		targetNameTypeMap = make(map[string]struct{})
		targets           []ReportTarget
	)
	for _, m := range data {
		key := fmt.Sprintf("%s-%s", m.TargetName, m.TargetType)
		if _, ok := targetNameTypeMap[key]; ok {
			continue
		}
		targets = append(targets, ReportTarget{TargetType: m.TargetType, TargetName: m.TargetName})
		targetNameTypeMap[key] = struct{}{}
	}
	err = ReportTargetCreateMulti(db, targets)
	if err != nil {
		err = fmt.Errorf("target create err: %w", err)
	}
	return
}
