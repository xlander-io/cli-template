package cmd_db

import (
	"github.com/coreservice-io/cli-template/basic/config"
	"github.com/coreservice-io/cli-template/plugin/g_counter_plugin"
	"github.com/coreservice-io/cli-template/plugin/sqldb_plugin"
	"github.com/coreservice-io/cli-template/src/common/dbkv"
	"github.com/coreservice-io/cli-template/src/user_mgr"
)

func Migrate() {
	StartDBComponent()

	//upgrade table
	sqldb_plugin.GetInstance().AutoMigrate(&user_mgr.UserModel{}, &dbkv.DBKVModel{})

	//try to init if enabled as database table may need to be created
	toml_conf := config.Get_config().Toml_config
	if toml_conf.General_counter.Enable {
		g_counter_plugin.DBMigrate(sqldb_plugin.GetInstance())
	}

}
