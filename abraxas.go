package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Response struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

const url = "https://api.groq.com/openai/v1/chat/completions"

func main() {
	key := flag.String("key", os.Getenv("GROQ_API_KEY"), "Groq API key")
	model := flag.String("model", "llama3-70b-8192", "model name")
	flag.Parse()
	if *key == "" {
		log.Fatal("GROQ_API_KEY not set")
	}
	c := &http.Client{}
	s := bufio.NewScanner(os.Stdin)
	m := []Message{}
	for {
		fmt.Print("abraxas> ")
		if !s.Scan() || strings.TrimSpace(s.Text()) == "q" {
			break
		}
		uc := strings.TrimSpace(s.Text())
		m = append(m, Message{"user", uc})
		res, err := sendRequest(c, *key, *model, m)
		if err != nil {
			log.Fatal(err)
		}
		if len(res.Choices) == 0 {
			break
		}
		ac := res.Choices[0].Message.Content
		fmt.Println(ac)
		fmt.Println()
		m = append(m, Message{"assistant", ac})
	}
}

func sendRequest(client *http.Client, key, model string, messages []Message) (*Response, error) {
	b, err := json.Marshal(Request{model, messages})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	var r Response
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}
	return &r, nil
}
