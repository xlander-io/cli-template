package config

import (
	"os"
	"strings"

	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/cli-template/plugin/logger_plugin"
	ilog "github.com/coreservice-io/log"
)

// 1.setup the config toml file
// 2.setup the basic working directory
// 3.setup the basic logger
// 4.return the real args
func ConfigBasic(toml_target string) []string {
	//////////init config/////////////
	real_args := []string{}
	for _, arg := range os.Args {
		arg_lower := strings.ToLower(arg)
		if strings.HasPrefix(arg_lower, "-conf=") || strings.HasPrefix(arg_lower, "--conf=") {
			toml_target = strings.TrimPrefix(arg_lower, "--conf=")
			toml_target = strings.TrimPrefix(toml_target, "-conf=")
			continue
		}
		real_args = append(real_args, arg)
	}

	os.Args = real_args
	conf_err := Init_config(toml_target)
	if conf_err != nil {
		panic(conf_err)
	}

	/////set build/////////
	build_mode_err := basic.SetMode(Get_config().Toml_config.Build.Mode)
	if build_mode_err != nil {
		panic(build_mode_err)
	}

	/////set up basic logger ///////
	logs_abs_path := basic.AbsPath("logs")
	if err := logger_plugin.Init(logs_abs_path); err != nil {
		panic(err)
	}

	basic.Logger = logger_plugin.GetInstance()
	loglevel := ilog.ParseLogLevel(Get_config().Toml_config.Log.Level)
	basic.Logger.SetLevel(loglevel)

	////////////////////////////////
	basic.Logger.Debugln("loglevel used:", ilog.LogLevelToTag(loglevel))
	//log basic info
	basic.Logger.Debugln("------------------------------------")
	basic.Logger.Debugln("working dir:", basic.WORK_DIR)
	basic.Logger.Debugln("------------------------------------")
	basic.Logger.Debugln("using user config toml file:", Get_config().User_config_path)
	///////////////////////////////

	return real_args

}
