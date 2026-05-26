package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk"
	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/ardanlabs/kronk/sdk/tools/defaults"
	"github.com/ardanlabs/kronk/sdk/tools/libs"
	"github.com/ardanlabs/kronk/sdk/tools/models"
	"github.com/hybridgroup/yzma/pkg/download"
	"github.com/labstack/echo/v4"
)

const (
	modelSource    = "unsloth/Qwen3-0.6B-Q8_0"
	requestTimeout = 120 * time.Second
)

type application struct {
	krn *kronk.Kronk
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("\nERROR: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	lp, mp, err := installSystem()
	if err != nil {
		return fmt.Errorf("unable to installation system: %w", err)
	}

	krn, err := newKronk(lp, mp)
	if err != nil {
		return fmt.Errorf("unable to init kronk: %w", err)
	}

	defer func() {
		fmt.Println("\nUnloading Kronk")
		if err := krn.Unload(context.Background()); err != nil {
			fmt.Printf("failed to unload model: %v", err)
		}
	}()

	app := application{krn: krn}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Kronk API running",
		})
	})
	e.GET("/api/hello", app.handlePrompt("Hello, how are you?"))
	e.GET("/api/bye", app.handlePrompt("Goodbye"))

	serverAddress := listenAddress()
	fmt.Printf("listening on http://localhost%s\n", serverAddress)
	return e.Start(serverAddress)
}

func installSystem() (string, models.Path, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	basePath := kronkBasePath()

	libMgr, err := libs.New(
		libs.WithBasePath(basePath),
		libs.WithVersion(defaults.LibVersion("")),
		libs.WithProcessor(download.CPU),
	)
	if err != nil {
		return "", models.Path{}, err
	}

	if _, err := libMgr.Download(ctx, kronk.FmtLogger); err != nil {
		return "", models.Path{}, fmt.Errorf("unable to install llama.cpp: %w", err)
	}

	mdls, err := models.NewWithPaths(basePath)
	if err != nil {
		return "", models.Path{}, fmt.Errorf("unable to init models: %w", err)
	}

	mp, err := mdls.Download(ctx, kronk.FmtLogger, modelSource)
	if err != nil {
		return "", models.Path{}, fmt.Errorf("unable to install model: %w", err)
	}

	return libMgr.LibsPath(), mp, nil
}

func newKronk(libPath string, mp models.Path) (*kronk.Kronk, error) {
	fmt.Println("loading model...")

	if err := kronk.Init(kronk.WithLibPath(libPath)); err != nil {
		return nil, fmt.Errorf("unable to init kronk: %w", err)
	}

	krn, err := kronk.New(
		model.WithModelFiles(mp.ModelFiles),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create inference model: %w", err)
	}

	return krn, nil
}

func (app application) handlePrompt(prompt string) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), requestTimeout)
		defer cancel()

		answer, err := askModel(ctx, app.krn, prompt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"prompt":   prompt,
			"response": answer,
		})
	}
}

func askModel(ctx context.Context, krn *kronk.Kronk, prompt string) (string, error) {
	d := model.D{
		"messages": model.DocumentArray(
			model.TextMessage(model.RoleUser, prompt),
		),
		"temperature": 0.7,
		"top_p":       0.9,
		"top_k":       40,
		"max_tokens":  2048,
	}

	resp, err := krn.Chat(ctx, d)
	if err != nil {
		return "", fmt.Errorf("chat: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("chat: empty response")
	}

	choice := resp.Choices[0]
	switch {
	case choice.Message != nil:
		return choice.Message.Content, nil
	case choice.Delta != nil:
		return choice.Delta.Content, nil
	default:
		return "", fmt.Errorf("chat: missing message content")
	}
}

func listenAddress() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return ":" + port
}

func kronkBasePath() string {
	basePath := os.Getenv("KRONK_BASE_PATH")
	if basePath == "" {
		basePath = defaults.BaseDir("")
	}

	return basePath
}
