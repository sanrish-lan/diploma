package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/kluctl/go-embed-python/python"
)

var ep *python.EmbeddedPython

const inlineBridge = `
import sys, json, vdom_app

raw = sys.stdin.read()
if raw:
    req = json.loads(raw)
    res = vdom_app.process_request(req)
    sys.stdout.write(json.dumps(res))
`

type VdomResponse struct {
	Status      int    `json:"status"`
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[k] = v[0]
	}

	payload, _ := json.Marshal(map[string]any{
		"method":  r.Method,
		"path":    r.URL.RequestURI(),
		"headers": headers,
	})

	pwd, _ := os.Getwd()
	cmd, err := ep.PythonCmd("-c", inlineBridge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cmd.Dir = pwd
	cmd.Stdin = bytes.NewReader(payload)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("[Python Error]: %s", stderr.String())
		http.Error(w, "500 Internal Python Error", http.StatusInternalServerError)
		return
	}

	var resp VdomResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		http.Error(w, "JSON Parse Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", resp.ContentType)
	w.WriteHeader(resp.Status)
	io.WriteString(w, resp.Body)
}

func main() {
	log.Println("[Go] Распаковка и запуск встроенного Python...")
	var err error
	ep, err = python.NewEmbeddedPython("vdom-app-runtime")
	if err != nil {
		log.Fatalf("Ошибка распаковки: %v", err)
	}

	log.Println("[Go] Сервер готов на http://localhost:8080")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
