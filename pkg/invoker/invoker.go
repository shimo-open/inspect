package invoker

import (
	"github.com/ego-component/egorm"
)

var (
	Db *egorm.Component
)

func Init() error {
	Db = egorm.Load("mysql").Build()
	return nil
}
