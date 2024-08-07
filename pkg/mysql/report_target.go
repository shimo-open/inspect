package mysql

import (
	"fmt"
	"time"

	"github.com/gotomicro/cetus/l"
	"github.com/gotomicro/ego/core/elog"
	"gorm.io/gorm"
)

type ReportTarget struct {
	Model
	ID         int64  `gorm:"auto;not null;comment:主键ID" json:"id" form:"id"`                                // 主键ID
	TargetName string `gorm:"type:varchar(255);not null;comment:关联目标名称" json:"targetName" form:"targetName"` // 关联目标名
	TargetType string `gorm:"type:varchar(255);not null;comment:关联目标类型" json:"targetType" form:"targetType"` // 关联目标类型
}

func (ReportTarget) TableName() string {
	return "report_target"
}

// ReportTargetCreateMulti 创建一条记录
func ReportTargetCreateMulti(db *gorm.DB, data []ReportTarget) (err error) {
	if len(data) == 0 {
		return nil
	}
	t := time.Now().Unix()
	q := "INSERT INTO `report_target` (`target_name`,`target_type`,`ctime`,`utime`,`dtime`) VALUES "
	inserts := make([]interface{}, 0, 4*len(data))
	for _, d := range data {
		q += fmt.Sprintf("(?,?,?,?,0),")
		inserts = append(inserts, d.TargetName, d.TargetType, t, t)
	}
	q = q[:len(q)-1]
	q += " ON DUPLICATE KEY UPDATE `ctime`=`ctime`"
	if err = db.Model(&ReportTarget{}).
		Set("gorm2dm:on_conflict_columns", []string{"target_name", "target_type"}).
		Exec(q, inserts...).Error; err != nil {
		elog.Warn("ReportTargetCreateMulti error", l.E(err))
		err = fmt.Errorf("target create err: %w", err)
		return
	}
	return
}
