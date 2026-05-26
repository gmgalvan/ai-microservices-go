package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk"
	"github.com/ardanlabs/kronk/sdk/tools/defaults"
	"github.com/ardanlabs/kronk/sdk/tools/libs"
	"github.com/ardanlabs/kronk/sdk/tools/models"
	"github.com/hybridgroup/yzma/pkg/download"
)

const modelSource = "unsloth/Qwen3-0.6B-Q8_0"

func main() {
	if err := run(); err != nil {
		fmt.Printf("\nERROR: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	basePath := kronkBasePath()

	libMgr, err := libs.New(
		libs.WithBasePath(basePath),
		libs.WithVersion(defaults.LibVersion("")),
		libs.WithProcessor(download.CPU),
	)
	if err != nil {
		return fmt.Errorf("init libs: %w", err)
	}

	if _, err := libMgr.Download(ctx, kronk.FmtLogger); err != nil {
		return fmt.Errorf("download llama.cpp: %w", err)
	}

	mdls, err := models.NewWithPaths(basePath)
	if err != nil {
		return fmt.Errorf("init models: %w", err)
	}

	if _, err := mdls.Download(ctx, kronk.FmtLogger, modelSource); err != nil {
		return fmt.Errorf("download model: %w", err)
	}

	fmt.Printf("preload complete at %s\n", basePath)
	return nil
}

func kronkBasePath() string {
	basePath := os.Getenv("KRONK_BASE_PATH")
	if basePath == "" {
		basePath = "/opt/kronk"
	}

	return basePath
}
