package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/prefeitura-rio/app-go-api/internal/auth"
	"github.com/prefeitura-rio/app-go-api/internal/config"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create token manager
	tokenManager := auth.NewServiceAccountTokenManager(
		cfg.Keycloak.URL,
		cfg.Keycloak.Realm,
		cfg.Keycloak.ClientID,
		cfg.Keycloak.ClientSecret,
	)

	// Get token
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := tokenManager.GetToken(ctx)
	if err != nil {
		log.Fatalf("Failed to get service account token: %v", err)
	}

	// Print just the token so it can be used in scripts
	fmt.Println(token)

	// Print info to stderr
	fmt.Fprintf(os.Stderr, "\n✓ Service account token retrieved successfully\n")
	fmt.Fprintf(os.Stderr, "To use: export SERVICE_TOKEN=\"$(go run scripts/get-service-token.go)\"\n\n")
}
