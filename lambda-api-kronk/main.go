package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk"
	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/ardanlabs/kronk/sdk/tools/defaults"
	"github.com/ardanlabs/kronk/sdk/tools/libs"
	"github.com/ardanlabs/kronk/sdk/tools/models"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hybridgroup/yzma/pkg/download"
)

const (
	modelSource    = "unsloth/Qwen3-0.6B-Q8_0"
	requestTimeout = 120 * time.Second
)

var (
	initOnce sync.Once
	initErr  error
	app      application
)

type application struct {
	krn *kronk.Kronk
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if err := initialize(); err != nil {
		return jsonResponse(500, fmt.Sprintf(`{"error":%q}`, err.Error())), nil
	}

	switch req.RawPath {
	case "/", "":
		return jsonResponse(200, `{"message":"Kronk Lambda API running"}`), nil

	case "/api/hello":
		return app.handlePrompt(ctx, "Hello, how are you?")

	case "/api/bye":
		return app.handlePrompt(ctx, "Goodbye")

	default:
		return jsonResponse(404, `{"error":"not found"}`), nil
	}
}

func initialize() error {
	initOnce.Do(func() {
		lp, mp, err := installSystem()
		if err != nil {
			initErr = fmt.Errorf("install system: %w", err)
			return
		}

		krn, err := newKronk(lp, mp)
		if err != nil {
			initErr = fmt.Errorf("init kronk: %w", err)
			return
		}

		app = application{krn: krn}
	})

	return initErr
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
		return "", models.Path{}, fmt.Errorf("download llama.cpp: %w", err)
	}

	mdls, err := models.NewWithPaths(basePath)
	if err != nil {
		return "", models.Path{}, fmt.Errorf("init models: %w", err)
	}

	mp, err := mdls.Download(ctx, kronk.FmtLogger, modelSource)
	if err != nil {
		return "", models.Path{}, fmt.Errorf("download model: %w", err)
	}

	return libMgr.LibsPath(), mp, nil
}

func newKronk(libPath string, mp models.Path) (*kronk.Kronk, error) {
	if err := kronk.Init(kronk.WithLibPath(libPath)); err != nil {
		return nil, fmt.Errorf("kronk init: %w", err)
	}

	krn, err := kronk.New(
		model.WithModelFiles(mp.ModelFiles),
	)
	if err != nil {
		return nil, fmt.Errorf("create inference model: %w", err)
	}

	return krn, nil
}

func (app application) handlePrompt(ctx context.Context, prompt string) (events.APIGatewayV2HTTPResponse, error) {
	reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	answer, err := askModel(reqCtx, app.krn, prompt)
	if err != nil {
		return jsonResponse(500, fmt.Sprintf(`{"error":%q}`, err.Error())), nil
	}

	body := fmt.Sprintf(`{"prompt":%q,"response":%q}`, prompt, answer)
	return jsonResponse(200, body), nil
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
		return strings.TrimSpace(choice.Message.Content), nil
	case choice.Delta != nil:
		return strings.TrimSpace(choice.Delta.Content), nil
	default:
		return "", fmt.Errorf("chat: missing message content")
	}
}

func kronkBasePath() string {
	basePath := os.Getenv("KRONK_BASE_PATH")
	if basePath == "" {
		basePath = "/tmp/kronk"
	}

	return basePath
}

func jsonResponse(statusCode int, body string) events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"content-type": "application/json",
		},
		Body: body,
	}
}
