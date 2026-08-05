package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/ozontech/seq-ui/internal/app/config/loader"
	"github.com/ozontech/seq-ui/internal/app/config/migrate"
	v1 "github.com/ozontech/seq-ui/internal/app/config/v1"
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
		logger.Fatal("missing required parameter", zap.String("param", "-source"))
	}

	stat, err := os.Stat(source)
	if err != nil {
		logger.Fatal("stat source", zap.String("source", source), zap.Error(err))
	}

	sourceCfg, err := os.ReadFile(source) //nolint:gosec
	if err != nil {
		logger.Fatal("read source config", zap.String("source", source), zap.Error(err))
	}

	version, err := loader.ReadVersion(sourceCfg)
	if err != nil {
		logger.Fatal("read config version", zap.Error(err))
	}
	if version != loader.V1 {
		logger.Fatal(fmt.Sprintf("source config is not v%d, nothing to migrate", loader.V1), zap.Int("version", version))
	}

	backupPath := strings.TrimSuffix(source, filepath.Ext(source)) + ".bck" + filepath.Ext(source)
	if _, err := os.Stat(backupPath); err == nil {
		logger.Fatal("backup already exists, remove it before re-running", zap.String("backup", backupPath))
	}
	if err := os.WriteFile(backupPath, sourceCfg, stat.Mode()); err != nil {
		logger.Fatal("write backup config", zap.String("backup", backupPath), zap.Error(err))
	}

	var cfg v1.Config
	decoder := yaml.NewDecoder(bytes.NewReader(sourceCfg))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		logger.Fatal("parse config v1", zap.Error(err))
	}

	migratedCfg := migrate.V1ToV2(cfg)

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&migratedCfg); err != nil {
		logger.Fatal("encode config v2", zap.Error(err))
	}
	if err := encoder.Close(); err != nil {
		logger.Fatal("close encoder", zap.Error(err))
	}

	if err := os.WriteFile(source, buf.Bytes(), stat.Mode()); err != nil {
		logger.Fatal("write migrated config", zap.Error(err))
	}

	logger.Info("config migrated", zap.String("source", source), zap.String("backup", backupPath))
}
