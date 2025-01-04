package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AnkiRequest struct {
	Action  string      `json:"action"`
	Version int         `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

type AnkiResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

type Anki struct {
	connectURL string
}

func NewAnki(connectURL string) *Anki {
	return &Anki{connectURL: connectURL}
}

func (a *Anki) invoke(action string, params interface{}) (interface{}, error) {
	request := AnkiRequest{
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

	var ankiResp AnkiResponse
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

func (a *Anki) AddNote(deckName, modelName string, fields map[string]string, tags []string) error {
	note := map[string]interface{}{
		"deckName":  deckName,
		"modelName": modelName,
		"fields":    fields,
		"tags":      tags,
		"options":   map[string]bool{"allowDuplicate": false},
	}

	_, err := a.invoke("addNote", map[string]interface{}{"note": note})
	if err != nil {
		return fmt.Errorf("failed to add note: %v", err)
	}

	fmt.Printf("Note added to deck '%s'\n", deckName)
	return nil
}
