package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/coreservice-io/cli-template/basic/config"
	"github.com/coreservice-io/cli-template/cmd_conf"
	"github.com/coreservice-io/cli-template/cmd_db"
	"github.com/coreservice-io/cli-template/cmd_default"
	"github.com/coreservice-io/cli-template/cmd_default/http/api"
	"github.com/coreservice-io/cli-template/cmd_log"
)

const CMD_NAME_DEFAULT = "default"
const CMD_NAME_GEN_API = "gen_api"
const CMD_NAME_LOG = "log"
const CMD_NAME_DB = "db"
const CMD_NAME_CONFIG = "config"

// //////config to do cmd ///////////
func ConfigCmd() *cli.App {

	real_args := config.ConfigBasic("default")

	var defaultAction = func(clictx *cli.Context) error {
		cmd_default.StartDefault()
		return nil
	}

	if len(real_args) > 1 {
		defaultAction = nil
	}

	return &cli.App{
		Action: defaultAction, //only run if no sub command

		//run if sub command not correct
		CommandNotFound: func(context *cli.Context, s string) {
			fmt.Println("command not find, use -h or --help show help")
		},

		Commands: []*cli.Command{
			{
				Name:  CMD_NAME_GEN_API,
				Usage: "api command",
				Action: func(clictx *cli.Context) error {
					api.GenApiDocs()
					return nil
				},
			},
			{
				Name:  CMD_NAME_LOG,
				Usage: "print all logs",
				Flags: cmd_log.GetFlags(),
				Action: func(clictx *cli.Context) error {
					num := clictx.Int64("num")
					onlyerr := clictx.Bool("only_err")
					cmd_log.StartLog(onlyerr, num)
					return nil
				},
			},
			{
				Name:  CMD_NAME_DB,
				Usage: "db command",
				Subcommands: []*cli.Command{
					{
						Name:  "init",
						Usage: "initialize db data",
						Action: func(clictx *cli.Context) error {
							fmt.Println("======== start of db data initialization ========")
							cmd_db.Initialize()
							fmt.Println("======== end  of  db data initialization ========")
							return nil
						},
					},
				},
			},
			{
				Name:  CMD_NAME_CONFIG,
				Usage: "config command",
				Subcommands: []*cli.Command{
					//show config
					{
						Name:  "show",
						Usage: "show configs",
						Action: func(clictx *cli.Context) error {
							fmt.Println("======== start of config ========")
							configs, _ := config.Get_config().Read_merge_config()
							fmt.Println(configs)
							fmt.Println("======== end  of  config ========")
							return nil
						},
					},
					//set config
					{
						Name:  "set",
						Usage: "set config",
						Flags: append(cmd_conf.Cli_get_flags(), &cli.StringFlag{Name: "config", Required: false}),
						Action: func(clictx *cli.Context) error {
							return cmd_conf.Cli_set_config(clictx)
						},
					},
				},
			},
		},
	}
}
