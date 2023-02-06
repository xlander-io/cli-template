package cmd_db

import (
	"errors"

	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/plugin/sqldb_plugin"
	"github.com/coreservice-io/cli-template/src/common/dbkv"
	"github.com/coreservice-io/cli-template/src/common/smart_cache"
	"github.com/coreservice-io/cli-template/src/user_mgr"
	"gorm.io/gorm"
)

// =====below data can be changed=====
var ini_admin_email = "admin@coreservice.com"
var ini_admin_password = "to_be_reset"
var ini_admin_roles = user_mgr.UserRoles
var ini_admin_permissions = user_mgr.UserPermissions

func Initialize() {
	StartDBComponent()

	key := "db_initialized"

	err := sqldb_plugin.GetInstance().Transaction(func(tx *gorm.DB) error {

		_, err := dbkv.GetDBKV(tx, nil, &key, false, false)

		if err == nil {
			return errors.New("db already initialized")
		}

		if err != smart_cache.QueryNilErr {
			return errors.New("db error:" + err.Error())
		}

		// create your own data here which won't change in the future
		configAdmin(tx)

		// dbkv
		return dbkv.SetDBKV_Bool(tx, key, true, "db initialized mark")
	})

	if err != nil {
		basic.Logger.Errorln("db initialize error:", err)
		return
	} else {
		basic.Logger.Infoln("db initialized")
	}

}

func configAdmin(tx *gorm.DB) {
	default_u_admin, err := user_mgr.CreateUser(tx, ini_admin_email, ini_admin_password, true, ini_admin_roles, ini_admin_permissions, "127.0.0.1")
	if err != nil {
		basic.Logger.Panicln(err)
	} else {
		basic.Logger.Infoln("default admin created:", default_u_admin)
	}
}
