package cmd_db

import (
	"github.com/coreservice-io/cli-template/plugin/sqldb_plugin"
	"github.com/coreservice-io/cli-template/src/common/dbkv"
	"github.com/coreservice-io/cli-template/src/user_mgr"
)

func Migrate() {
	StartDBComponent()

	//upgrade table
	sqldb_plugin.GetInstance().AutoMigrate(&user_mgr.UserModel{}, &dbkv.DBKVModel{})
}
