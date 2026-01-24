# Anki LLM

This project provides a command-line tool for generating Anki flashcards from the contents of a PDF file. It leverages
the Gemini LLM (Language Model) to process the PDF content and create meaningful flashcards, which are then added to an
Anki deck using the AnkiConnect API.

> **Note:** This project is a work in progress and is not yet ready for general use.

## Installation

1. Install the AnkiConnect plugin in Anki. You can find the plugin on the AnkiWeb
   site [here](https://ankiweb.net/shared/info/2055492159).
2. Install Go on your machine. You can find the installation instructions [here](https://golang.org/doc/install).
3. Add `GEMINI_API_KEY` to your environment variables. You can find your API key
   at [Google AI Studio]("https://aistudio.google.com/").
2. Clone the repository

  ```bash
  git clone https://github.com/sotterbeck/anki-llm.git
  ```

## Usage

To generate Anki flashcards from a PDF file, run the following command:

```bash
go run ./cmd/anki-llm
```

> [!CAUTION]
> While this tool automates the process of generating flashcards from a PDF, it’s important to recognize that simply
> converting content from a document into flashcards without thoughtful engagement may not be the most effective way to
> learn. Please use this tool with caution and consider how you can best leverage it to support your learning goals.
