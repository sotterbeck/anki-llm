package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

// Main entry point of the application
func main() {
	godotenv.Load()
	pdfFile, deckName := parseArguments()

	file := openPDFFile(pdfFile)
	defer file.Close()

	ctx := context.Background()
	modelName := "Basic"
	anki := initializeAnkiClient()
	llm := initializeLLM(ctx)
	defer llm.Close()

	if err := anki.CreateDeck(deckName); err != nil {
		log.Fatalf("Error creating or accessing deck: %v", err)
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Generating Anki notes..."

	s.Start()
	notes, err := llm.GenerateAnkiNotes(ctx, file, modelName)
	if err != nil {
		log.Fatalf("Failed to generate Anki notes: %v", err)
	}
	s.Stop()

	fmt.Printf("Generated notes:\n")
	for _, note := range notes {
		fmt.Printf(" - %s\n", note["Front"])
	}

	addNotesToDeck(anki, deckName, modelName, notes)

	fmt.Printf("Successfully added %d notes to the '%s' deck.\n", len(notes), deckName)
}

// Parse CLI arguments and return the PDF file path and deck name
func parseArguments() (string, string) {
	pdfFile := flag.String("pdf", "", "Path to the PDF file to process (required)")
	deckName := flag.String("deck", "", "Name of the Anki deck to create or use (required)")
	flag.Parse()

	// Validate that both arguments are provided
	if *pdfFile == "" || *deckName == "" {
		fmt.Println("Error: Both '-pdf' and '-deck' arguments are required.")
		flag.Usage()
		os.Exit(1)
	}

	return *pdfFile, *deckName
}

// Open the specified PDF file and return the file handle
func openPDFFile(pdfFile string) *os.File {
	file, err := os.Open(pdfFile)
	if err != nil {
		log.Fatalf("Failed to open PDF file: %v", err)
	}
	return file
}

// Initialize and return the Anki client
func initializeAnkiClient() *Anki {
	return NewAnki("http://localhost:8765")
}

// Initialize and return the Gemini LLM client
func initializeLLM(ctx context.Context) LLM {
	llm, err := NewGeminiLLM(ctx, "gemini-2.0-flash-exp", os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		log.Fatalf("Failed to create Gemini LLM: %v", err)
	}
	return llm
}

// Add notes to the specified Anki deck
func addNotesToDeck(anki *Anki, deckName string, modelName string, notes []map[string]string) {
	if err := anki.AddNotes(deckName, modelName, notes); err != nil {
		log.Fatalf("Failed to add notes to deck: %v", err)
	}
}
