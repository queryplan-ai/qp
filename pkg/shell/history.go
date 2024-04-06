package shell

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/queryplan-ai/qp/pkg/shell/types"
)

const (
	maxCommandHistory = 100
)

func getHistoryFilePath() (string, error) {
	stateDir := filepath.Join(xdg.StateHome, "qp")

	// Ensure the directory exists
	if err := os.MkdirAll(stateDir, os.ModePerm); err != nil {
		return "", err
	}

	// Return the full path to the history file
	return filepath.Join(stateDir, "history"), nil
}

func trimHistory(sh *types.Shell) error {
	file, err := os.Open(sh.HistoryFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) > sh.HistoryMaxSize {
		lines = lines[len(lines)-sh.HistoryMaxSize:]
	}

	file.Close()

	err = os.WriteFile(sh.HistoryFilePath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
	if err != nil {
		return err
	}

	return nil
}
