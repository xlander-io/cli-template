package cmd_log

import (
	"github.com/coreservice-io/cli-template/basic"
	"github.com/coreservice-io/log"
)

func StartLog(onlyerr bool, num int64) {
	if num == 0 {
		num = 20
	}
	if onlyerr {
		basic.Logger.PrintLastN(num, []log.LogLevel{log.PanicLevel, log.FatalLevel, log.ErrorLevel})
	} else {
		basic.Logger.PrintLastN(num, []log.LogLevel{log.PanicLevel, log.FatalLevel, log.ErrorLevel, log.InfoLevel, log.WarnLevel, log.DebugLevel, log.TraceLevel})
	}
}
