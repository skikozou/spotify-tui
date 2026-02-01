package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"spotify-tui/internal/auth"
	"spotify-tui/internal/config"
	"spotify-tui/internal/logger"
	"spotify-tui/internal/spotify"
	"spotify-tui/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const defaultLogFile = "log/spotify-tui.log"

func main() {
	// Parse command-line flags
	debug := flag.Bool("debug", false, "Enable debug mode (log to file)")
	logFile := flag.String("log-file", "", "Log file path (default: ./log/spotify-tui.log)")
	flag.Parse()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logFilePath := *logFile
	if logFilePath == "" {
		logFilePath = defaultLogFile
	}
	if err := logger.Init(*debug, logFilePath); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Application started", "debug", *debug)

	// Check if client credentials are set
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		fmt.Println("Spotify credentials not found.")
		fmt.Println("Please set your Spotify Client ID and Client Secret:")
		fmt.Println()
		fmt.Println("1. Go to https://developer.spotify.com/dashboard")
		fmt.Println("2. Create an application")
		fmt.Println("3. Set redirect URI to: http://localhost:8080/callback")
		fmt.Println("4. Copy your Client ID and Client Secret")
		fmt.Println()

		fmt.Print("Enter Client ID: ")
		fmt.Scanln(&cfg.ClientID)
		fmt.Print("Enter Client Secret: ")
		fmt.Scanln(&cfg.ClientSecret)

		if err := cfg.Save(); err != nil {
			log.Fatalf("Failed to save config: %v", err)
		}
	}

	// Authenticate
	spotifyClient, err := auth.Authenticate(cfg)
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}

	// Create client wrapper
	client := spotify.NewClient(spotifyClient)

	// Create Bubbletea model
	ctx := context.Background()
	model := ui.NewModel(ctx, client)

	// Start TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
