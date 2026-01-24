package anki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	Action  string      `json:"action"`
	Version int         `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

type Response struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

type Anki struct {
	connectURL string
}

type Note struct {
	DeckName  string            `json:"deckName"`
	ModelName string            `json:"modelName"`
	Fields    map[string]string `json:"fields"`
	Tags      []string          `json:"tags"`
	Options   map[string]bool   `json:"options"`
}

func NewAnki(connectURL string) *Anki {
	return &Anki{connectURL: connectURL}
}

func (a *Anki) invoke(action string, params interface{}) (interface{}, error) {
	request := Request{
		Action:  action,
		Version: 6,
		Params:  params,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(a.connectURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var ankiResp Response
	if err := json.Unmarshal(body, &ankiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if ankiResp.Error != "" {
		return nil, fmt.Errorf("error from AnkiConnect: %s", ankiResp.Error)
	}

	return ankiResp.Result, nil
}

func (a *Anki) CreateDeck(deckName string) error {
	_, err := a.invoke("createDeck", map[string]string{"deck": deckName})
	if err != nil {
		return fmt.Errorf("failed to create deck: %v", err)
	}
	fmt.Printf("Deck '%s' created successfully\n", deckName)
	return nil
}

func (a *Anki) ListDeckNames() ([]string, error) {
	result, err := a.invoke("deckNames", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck names: %v", err)
	}

	slice, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	names := make([]string, len(slice))
	for i, v := range slice {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected element type at index %d: %T", i, v)
		}
		names[i] = str
	}

	return names, nil
}

func (a *Anki) AddNotes(deckName, modelName string, notes []map[string]string) error {
	var noteData []Note
	for _, note := range notes {
		noteData = append(noteData, Note{
			DeckName:  deckName,
			ModelName: modelName,
			Fields:    note,
			Options:   map[string]bool{"allowDuplicate": false},
		})
	}

	_, err := a.invoke("addNotes", map[string]interface{}{"notes": noteData})
	if err != nil {
		return fmt.Errorf("failed to add notes: %v", err)
	}

	return nil
}
