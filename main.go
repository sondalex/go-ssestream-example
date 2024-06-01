package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.New("index.html").ParseFiles("index.html")
	if err != nil {
		panic(err)
	}
}

func MakeRootHandler(ch chan string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			err := tmpl.Execute(w, nil)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		case "POST":
			partial, err := template.New("partial.html").ParseFiles("partial.html")
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			go func() {
				for i := 0; i < 10; i++ {
					token := strings.Join([]string{"token", strconv.Itoa(i)}, "_")
					ch <- strings.Join([]string{token, " "}, "")
					time.Sleep(1 * time.Second)
				}
			}()
			partial.Execute(w, nil)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func MakeSSEHandler(ch chan string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		notify := w.(http.CloseNotifier).CloseNotify()
		flusher, ok := w.(http.Flusher)

		if !ok {
			http.Error(w, "Not Flushed", http.StatusInternalServerError)
			return
		}

		for {
			select {
			case <-notify:
				fmt.Println("Client has closed the connection")
				return
			case text := <-ch:
				sseText := fmt.Sprintf("data: %s\n\n", text)
				fmt.Fprint(w, sseText)
				flusher.Flush()
			case <-r.Context().Done():
				fmt.Println("Closed connection")
				return
			}
		}
	}
}

func main() {
	var ch chan string = make(chan string)
	http.HandleFunc("/", MakeRootHandler(ch))
	http.HandleFunc("/sse", MakeSSEHandler(ch))
	http.ListenAndServe(":1313", nil)
}
