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
			viper.BindEnv("openai-api-key", "OPENAI_KEY")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.GetViper()

			opts := shelltypes.ShellOpts{
				ConnectionURI: v.GetString("db-uri"),
				OpenAIAPIKey:  v.GetString("openai-api-key"),
			}

			// parse the args, args[0] should be the connection string, but it's optional
			// and overrides the env var if one one set
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

	cmd.Flags().String("db-uri", "", "database connection URI to automatically use")
	cmd.Flags().String("openai-api-key", "", "OpenAI API key to use")

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
	viper.SetEnvPrefix("QP")
	viper.AutomaticEnv()
}
