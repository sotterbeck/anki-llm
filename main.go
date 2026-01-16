package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/sotterbeck/anki-llm/ui"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go handleSignals(cancel)

	llm := initializeLLM(ctx, os.Getenv("GEMINI_MODEL"))
	defer llm.Close()
	anki := initializeAnkiClient()

	// Create and start the TUI program
	uiModel := ui.NewModel(ctx, llm, anki)
	p := tea.NewProgram(uiModel, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatalf("failed to start TUI: %v", err)
	}
}

// handleSignals cancels the provided context when the process receives SIGINT or SIGTERM.
func handleSignals(cancel context.CancelFunc) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	cancel()
}

// initializeAnkiClient returns an Anki client. Adjust the URL here if needed.
func initializeAnkiClient() *Anki {
	return NewAnki("http://localhost:8765")
}

// initializeLLM creates the Gemini LLM client. The model parameter may be empty;
// if so, a sensible default is chosen.
func initializeLLM(ctx context.Context, model string) LLM {
	if model == "" {
		model = "gemini-3-flash-preview"
	}
	llm, err := NewGeminiLLM(ctx, model, os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		log.Fatalf("Failed to create Gemini LLM: %v", err)
	}
	return llm
}
