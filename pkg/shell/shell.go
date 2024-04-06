package shell

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/queryplan-ai/qp/pkg/shell/types"
)

var (
	ErrExit = fmt.Errorf("exit")
)

func RunShell(opts types.ShellOpts) error {
	historyFile, err := getHistoryFilePath()
	if err != nil {
		return err
	}

	sh := types.Shell{
		HistoryFilePath: historyFile,
		HistoryMaxSize:  maxCommandHistory,
	}

	if opts.ConnectionURI != "" {
		result := handleConnect(&sh, opts.ConnectionURI)
		if !result.IsSuccess {
			return fmt.Errorf("error connecting to database: %s", result.Message)
		}
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          prompt(&sh),
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		log.Printf("Error creating readline: %v", err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}

		if err := trimHistory(&sh); err != nil {
			log.Printf("Error trimming history: ex%v", err)
		}

		result := processShellCommand(&sh, line)
		if result.IsFatal {
			if result.IsSuccess {
				os.Exit(0)
			}

			fmt.Printf("Error: %s\n", result.Message)
			os.Exit(1)
		} else {
			if !result.IsSuccess {
				fmt.Printf("Error: %s\n", result.Message)
			} else {
				if result.Message != "" {
					fmt.Println(result.Message)
				}
			}
		}

		// update the prompt
		rl.SetPrompt(prompt(&sh))
	}

	return nil
}

func prompt(sh *types.Shell) string {
	if sh.DB == nil {
		return "<not connected, use /connect> >>> "
	}

	return fmt.Sprintf("%s/%s >>> ", sh.DatabaseEngine, sh.DatabaseName)
}

func processShellCommand(sh *types.Shell, cmd string) *types.ShellCommandResult {
	if !strings.HasPrefix(cmd, "/") {
		return handleQuery(sh, stripCommand(cmd))
	}

	cmdParts := strings.Split(cmd, " ")
	switch cmdParts[0] {
	case "/exit":
		return &types.ShellCommandResult{
			IsSuccess: true,
			IsFatal:   true,
		}
	case "/help", "/?":
		return showHelp()
	case "/connect":
		return handleConnect(sh, stripCommand(cmd))
	default:
		return showUnknownCommand()
	}
}

func stripCommand(cmd string) string {
	// remove the / command from the cmd
	parts := strings.Fields(cmd)
	if strings.HasPrefix(parts[0], "/") {
		return strings.Join(parts[1:], " ")
	}

	return cmd
}

func showHelp() *types.ShellCommandResult {
	return &types.ShellCommandResult{
		IsSuccess: false,
		IsFatal:   false,
		Message:   "not implemented",
	}
}

func showUnknownCommand() *types.ShellCommandResult {
	return &types.ShellCommandResult{
		IsSuccess: false,
		IsFatal:   false,
		Message:   "not implemented",
	}
}
