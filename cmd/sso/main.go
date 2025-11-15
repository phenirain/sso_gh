package main

// @title SSO API
// @version 1.0
// @description SSO service API.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token in the format: Bearer {token}
import (
	"log/slog"
	"os"

	"github.com/phenirain/sso/internal"
	"github.com/phenirain/sso/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Could not load config", "err", err)
		os.Exit(1)
	}
	if err := internal.Run(cfg); err != nil {
		slog.Error("Failed to run server", "err", err)
		os.Exit(1)
	}
}
