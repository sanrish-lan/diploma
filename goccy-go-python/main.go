package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	python "github.com/goccy/go-python"
	"github.com/goccy/go-python/fs"
)

type VDOMServer struct {
	py *python.Python
	fn python.FunctionValue
}

func NewServer() (*VDOMServer, error) {
	ctx := context.Background()

	stdlibDir, err := fs.ExtractStdlib()
	if err != nil {
		return nil, fmt.Errorf("stdlib error: %w", err)
	}

	cfg := python.Config{
		FS:        fs.NewHostFS(),
		StdlibDir: stdlibDir,
	}

	p, err := python.New(cfg)
	if err != nil {
		return nil, err
	}

	pwd, _ := os.Getwd()
	bootstrap := fmt.Sprintf("import sys\nsys.path.insert(0, %q)\nimport vdom_app", pwd)

	if _, err := p.Eval(ctx, bootstrap); err != nil {
		return nil, fmt.Errorf("import vdom_app failed: %w", err)
	}

	fnVal, err := p.Eval(ctx, "vdom_app.process_request")
	if err != nil {
		return nil, err
	}

	fn, err := python.As[python.FunctionValue](fnVal.Value)
	if err != nil {
		return nil, fmt.Errorf("process_request is not callable: %w", err)
	}

	return &VDOMServer{py: p, fn: fn}, nil
}

func (s *VDOMServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var headerItems []python.DictItem
	for k, v := range r.Header {
		headerItems = append(headerItems, python.DictItem{
			Key:   python.ValueOf(k),
			Value: python.ValueOf(v[0]),
		})
	}

	headersDict, err := s.py.NewDict(ctx, headerItems...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reqDict, err := s.py.NewDict(ctx,
		python.DictItem{Key: python.ValueOf("method"), Value: python.ValueOf(r.Method)},
		python.DictItem{Key: python.ValueOf("path"), Value: python.ValueOf(r.URL.RequestURI())},
		python.DictItem{Key: python.ValueOf("headers"), Value: headersDict},
		python.DictItem{Key: python.ValueOf("engine"), Value: python.ValueOf("goccy/go-python (Pure-Go WebAssembly CPython 3.14)")},
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}


	resVal, err := s.fn.Call(ctx, reqDict)
	if err != nil {
		http.Error(w, fmt.Sprintf("Python call error: %v", err), http.StatusInternalServerError)
		return
	}

	resDict, err := python.As[python.DictValue](resVal)
	if err != nil {
		http.Error(w, "Response must be dict", http.StatusInternalServerError)
		return
	}

	statusVal, _, _ := resDict.Get(ctx, python.ValueOf("status"))
	statusInt, _ := python.As[python.IntValue](statusVal)
	statusCode, _ := statusInt.Int64()

	ctVal, _, _ := resDict.Get(ctx, python.ValueOf("content_type"))
	ctStr, _ := python.As[python.StrValue](ctVal)

	bodyVal, _, _ := resDict.Get(ctx, python.ValueOf("body"))
	bodyStr, _ := python.As[python.StrValue](bodyVal)

	w.Header().Set("Content-Type", ctStr.String())
	w.WriteHeader(int(statusCode))
	w.Write([]byte(bodyStr.String()))
}

func main() {
	log.Println("[Go] Инициализация Pure-Go Python 3.14 (без CGo)...")
	srv, err := NewServer()
	if err != nil {
		log.Fatalf("Ошибка запуска: %v", err)
	}
	defer srv.py.Close()

	log.Println("[Go] Сервер успешно запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", srv))
}
