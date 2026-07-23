package main

import (
	"errors"
	"flag"
	"os"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/logger"
	"go.uber.org/zap"
)

var (
	source = flag.String("source", "", "path to the config file to migrate in place")
)

func main() {
	flag.Parse()

	run(*source)
}

func run(source string) {
	if source == "" {
		logger.Fatal("missing required parameter", zap.String("param", "-source"))
	}

	legacyCfg, err := os.ReadFile(source)
	if err != nil {
		logger.Fatal("error reading legacy config", zap.String("source", source), zap.Error(err))
	}

	backupPath := source + ".bak"

	if _, err := os.Stat(backupPath); err == nil {
		logger.Fatal("backup already exists, remove it before re-running", zap.String("backup", backupPath))
	} else if !errors.Is(err, os.ErrNotExist) {
		logger.Fatal("stat backup", zap.String("backup", backupPath), zap.Error(err))
	}
	if err := os.WriteFile(backupPath, legacyCfg, 0644); err != nil {
		logger.Fatal("error writing backup legacy config", zap.String("backup", backupPath), zap.Error(err))
	}

	backupCfg, err := config.FromFile(backupPath)
	if err != nil {
		logger.Fatal("read backup config file error", zap.Error(err))
	}
}
