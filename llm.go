package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"io"
	"log"
)

var prompt = `
Extract key concepts, definitions, and questions from this PDF document to create a series of Anki flashcards, ensuring the cards are in the same language as the document. Each flashcard should consist of a concise question on the front and a detailed answer on the back. The questions should test understanding of the core concepts likely to appear in the exam. The format of the cards should be consistent:
 - Front: A clear, concise question related to the material in the PDF.
 - Back: A complete answer with important details, but not overly verbose.

Do not add any information that is not directly present in the document. Avoid hallucinating or introducing external knowledge. The language of the flashcards should match the language of the PDF.
`

type LLM interface {
	// GenerateAnkiNotes generates Anki notes from the given reader.
	// The reader is expected to contain the content to be converted to Anki notes.
	GenerateAnkiNotes(ctx context.Context, r io.Reader, noteModel string) ([]map[string]string, error)

	Close() error
}

type GeminiLLM struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewGeminiLLM(ctx context.Context, model string, apiKey string) (LLM, error) {
	cli, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create generative client: %v", err)
	}

	m := cli.GenerativeModel(model)
	m.ResponseMIMEType = "application/json"
	return &GeminiLLM{client: cli, model: m}, nil
}

var schemas = map[string]genai.Schema{
	"Basic": {
		Type: genai.TypeArray,
		Items: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"Front": {Type: genai.TypeString},
				"Back":  {Type: genai.TypeString},
			},
			Required: []string{"Front", "Back"},
		},
	},
}

func (g *GeminiLLM) GenerateAnkiNotes(ctx context.Context, r io.Reader, noteModel string) ([]map[string]string, error) {
	file, err := g.client.UploadFile(ctx, "", r, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %v", err)
	}
	defer g.client.DeleteFile(ctx, file.Name)

	schema, ok := schemas[noteModel]
	if !ok {
		return nil, fmt.Errorf("note model %q not found", noteModel)
	}
	g.model.ResponseSchema = &schema

	resp, err := g.model.GenerateContent(ctx,
		genai.Text(prompt),
		genai.FileData{URI: file.URI},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate anki card content: %v", err)
	}

	var notes []map[string]string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			if err := json.Unmarshal([]byte(txt), &notes); err != nil {
				log.Fatal(err)
			}
		}
	}
	return notes, nil
}

func (g *GeminiLLM) Close() error {
	return g.client.Close()
}
