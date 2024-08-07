package job

import (
	"fmt"

	"github.com/gotomicro/ego/task/ejob"
	"inspect/pkg/invoker"
	"inspect/pkg/mysql"
)

func RunInstall(ctx ejob.Context) error {
	models := []interface{}{
		&mysql.ReportMeasure{},
		&mysql.ReportTarget{},
	}
	gormdb := invoker.Db.Debug().WithContext(ctx.Ctx)
	err := gormdb.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(models...)
	if err != nil {
		return err
	}
	fmt.Println("create table ok")
	return nil
}
