package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/queryplan-ai/qp/pkg/shell"
	shelltypes "github.com/queryplan-ai/qp/pkg/shell/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "qp",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := shelltypes.ShellOpts{}

			// parse the args, args[0] should be the connection string, but it's optional
			if len(args) > 0 {
				opts.ConnectionURI = args[0]
			}

			if err := shell.RunShell(opts); err != nil {
				return err
			}

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	cmd.AddCommand(VersionCmd())

	cmd.PersistentFlags().String("log-level", "info", "log level")

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetEnvPrefix("QUERYPLAN")
	viper.AutomaticEnv()
}
