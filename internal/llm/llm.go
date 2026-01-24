package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

var prompt = `
You are an intelligent assistant designed to generate Anki notes from PDF documents. Your goal is to extract key information from the PDF and format it into effective Anki flashcards.

Output Requirements:
- Each note must have a "Front" and "Back."
- The "Front" should be a question, incomplete statement, or a term requiring a definition or explanation.
- The "Back" should provide a concise, accurate, and complete answer or explanation. Use examples or clarifications where helpful.

Guidelines:
1. Focus on sections titled "Key Concepts," "Summary," or bolded/highlighted content. Ignore references, footnotes, or content unlikely to appear on a flashcard.
2. Limit text to 20–30 words per side for clarity and memory efficiency.
3. Where applicable, include formulas, diagrams, or tables in the "Back" to enhance understanding.
4. If a topic requires multiple explanations or steps, create separate flashcards for each aspect to ensure focus and recall.
5. Generate notes in the SAME language as the source PDF, even if some parts of the PDF are in a different language. For example, if the PDF is primarily in German, create the notes in German.

Using LaTeX in Anki Cards:
- ALWAYS use LaTeX for mathematical formulas, scientific notations, Greek symbols, formal definitions or any content that requires precise formatting.
- Examples of LaTeX usage:
  - Mathematical formulas: "What is the formula for the area of a circle?" \\( A = \\pi r^2 \\)
  - Chemical equations: "What is the balanced equation for photosynthesis?" \\( 6CO_2 + 6H_2O \\rightarrow C_6H_{12}O_6 + 6O_2 \\)
  - Greek symbols: "What does \\( \alpha \\) represent in physics?" \\( \\alpha \\) typically represents the angular acceleration or a fine-structure constant.
  - Formal definitions: "What is public-key cryptography?" Public-key cryptography is a method that uses two keys, a public key \\( k_{pub} \\) for encryption and a private key \\( k_{priv} \\) for decryption.
  - Complex diagrams or symbols that are best represented using LaTeX commands.

Using Inline and Block LaTeX:
- **Inline LaTeX:** Use for short formulas or symbols embedded within sentences. Enclose them in \\( and \\).
  - Example: "What is the circumference of a circle?" Answer: "The circumference is calculated as \\( C = 2\\pi r \\)."
- **Block LaTeX:** Use for longer equations, complex formulas, or diagrams that require better visual separation. Enclose them in \\[ and \\].
  - Example: "What is the quadratic formula?" Answer: "The solution to \\( ax^2 + bx + c = 0 \\) is given by:

    \\[
    x = \\frac{-b \\pm \\sqrt{b^2 - 4ac}}{2a}
    \\]
    "
- Use inline formatting for concise expressions within text and block formatting for emphasis or detailed visual representations.

Your output should be formatted as:
- "Front": <text of the prompt/question>
- "Back": <text of the answer/explanation>

Generate concise, clear, and focused notes designed for effective learning.`

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
