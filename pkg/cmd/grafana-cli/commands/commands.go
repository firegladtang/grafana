package commands

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/cmd/grafana-cli/commands/datamigrations"
	"github.com/grafana/grafana/pkg/cmd/grafana-cli/logger"
	"github.com/grafana/grafana/pkg/cmd/grafana-cli/utils"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
)

func runDbCommand(command func(commandLine utils.CommandLine, sqlStore *sqlstore.SqlStore) error) func(context *cli.Context) {
	return func(context *cli.Context) {
		cmd := &utils.ContextCommandLine{Context: context}

		cfg := setting.NewCfg()

		cfg.Load(&setting.CommandLineArgs{
			Config:   cmd.String("config"),
			HomePath: cmd.String("homepath"),
			Args:     context.Args(),
		})

		cfg.LogConfigSources()

		engine := &sqlstore.SqlStore{}
		engine.Cfg = cfg
		engine.Bus = bus.GetBus()
		engine.Init()

		if err := command(cmd, engine); err != nil {
			logger.Errorf("\n%s: ", color.RedString("Error"))
			logger.Errorf("%s\n\n", err)

			cmd.ShowHelp()
			os.Exit(1)
		}

		logger.Info("\n\n")
	}
}

func runPluginCommand(command func(commandLine utils.CommandLine) error) func(context *cli.Context) {
	return func(context *cli.Context) {

		cmd := &utils.ContextCommandLine{Context: context}
		if err := command(cmd); err != nil {
			logger.Errorf("\n%s: ", color.RedString("Error"))
			logger.Errorf("%s %s\n\n", color.RedString("✗"), err)

			cmd.ShowHelp()
			os.Exit(1)
		}

		logger.Info("\nRestart grafana after installing plugins . <service grafana-server restart>\n\n")
	}
}

var pluginCommands = []cli.Command{
	{
		Name:   "install",
		Usage:  "install <plugin id> <plugin version (optional)>",
		Action: runPluginCommand(installCommand),
	}, {
		Name:   "list-remote",
		Usage:  "list remote available plugins",
		Action: runPluginCommand(listremoteCommand),
	}, {
		Name:   "list-versions",
		Usage:  "list-versions <plugin id>",
		Action: runPluginCommand(listversionsCommand),
	}, {
		Name:    "update",
		Usage:   "update <plugin id>",
		Aliases: []string{"upgrade"},
		Action:  runPluginCommand(upgradeCommand),
	}, {
		Name:    "update-all",
		Aliases: []string{"upgrade-all"},
		Usage:   "update all your installed plugins",
		Action:  runPluginCommand(upgradeAllCommand),
	}, {
		Name:   "ls",
		Usage:  "list all installed plugins",
		Action: runPluginCommand(lsCommand),
	}, {
		Name:    "uninstall",
		Aliases: []string{"remove"},
		Usage:   "uninstall <plugin id>",
		Action:  runPluginCommand(removeCommand),
	},
}

var dbCommandFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "homepath",
		Usage: "path to grafana install/home path, defaults to working directory",
	},
	cli.StringFlag{
		Name:  "config",
		Usage: "path to config file",
	},
}

var adminCommands = []cli.Command{
	{
		Name:   "reset-admin-password",
		Usage:  "reset-admin-password <new password>",
		Action: runDbCommand(resetPasswordCommand),
		Flags:  dbCommandFlags,
	},
	{
		Name:  "data-migration",
		Usage: "Runs a script that migrates or cleanups data in your db",
		Subcommands: []cli.Command{
			{
				Name:   "encrypt-datasource-passwords",
				Usage:  "Migrates passwords from unsecured fields to secure_json_data field. Return ok unless there is an error. Safe to execute multiple times.",
				Action: runDbCommand(datamigrations.EncryptDatasourcePaswords),
				Flags:  dbCommandFlags,
			},
		},
	},
}

var Commands = []cli.Command{
	{
		Name:        "plugins",
		Usage:       "Manage plugins for grafana",
		Subcommands: pluginCommands,
	},
	{
		Name:        "admin",
		Usage:       "Grafana admin commands",
		Subcommands: adminCommands,
	},
}
