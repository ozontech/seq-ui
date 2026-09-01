package main

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/internal/app/config/loader"
	"github.com/ozontech/seq-ui/logger"
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
		logger.Fatal("miss required parameter", zap.String("param", "-source"))
	}

	stat, err := os.Stat(source)
	if err != nil {
		logger.Fatal("stat source", zap.String("source", source), zap.Error(err))
	}

	sourceCfg, err := os.ReadFile(source) //nolint:gosec
	if err != nil {
		logger.Fatal("read source config", zap.String("source", source), zap.Error(err))
	}

	migratedCfg, err := loader.ToLatestVersion(sourceCfg)
	if errors.Is(err, loader.ErrAlreadyLatestConfigVersion) {
		logger.Info("source config is already at the latest schema version, nothing to migrate", zap.String("source", source))
		return
	} else if err != nil {
		logger.Fatal("migrate config", zap.String("source", source), zap.Error(err))
	}

	backupPath := strings.TrimSuffix(source, filepath.Ext(source)) + ".bck" + filepath.Ext(source)
	if _, err := os.Stat(backupPath); err == nil {
		logger.Fatal("backup already exists, remove it before re-running", zap.String("backup", backupPath))
	}
	if err := os.WriteFile(backupPath, sourceCfg, stat.Mode()); err != nil {
		logger.Fatal("write backup config", zap.String("backup", backupPath), zap.Error(err))
	}

	if err := os.WriteFile(source, migratedCfg, stat.Mode()); err != nil {
		logger.Fatal("write migrated config", zap.Error(err))
	}

	logger.Info("config migrated", zap.String("source", source), zap.String("backup", backupPath))
}
