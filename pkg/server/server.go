package server

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/psu/todo-service/proto"
	"github.com/rs/zerolog/log"
)

var gitHubUrl = "https://api.github.com/gists"
var clientID = "7a3f63ef8ef5de99a305"
var clientSecret = "f802c7c697be1ccfbabda2eafe7e37f1719334b9"
var redirectUrl = "http://127.0.0.1:4200"

func StartServer(host string, port string, profilerPort string) {
	// Create router
	r := chi.NewRouter()
	// Setup routes
	r.Options("/", OptionsHandler)

	r.Post("/auth", Authorize)

	r.Get("/tasks", GetAllTask)

	r.Get("/tasks/{id}", GetTask)
	r.Post("/tasks/{id}", InsertTask)
	r.Put("/task/{id}", UpdateTaskStatus)
	r.Delete("/tasks/{id}", RemoveTask)

	// File routes
	r.Get("/tasks.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/js/tasks.js")
	})
	r.Get("/task.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/js/task.js")
	})
	r.Get("/auth.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/js/auth.js")
	})

	r.Get("/tasks.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/style/tasks.css")
	})
	r.Get("/task.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/style/task.css")
	})
	r.Get("/auth.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/style/auth.css")
	})

	r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/html/auth.gohtml")
	})

	// Server definition
	srv := http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, port),
		Handler: r,
	}

	// Graceful shutdown
	done := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM)
		signal.Notify(sigint, syscall.SIGINT)
		<-sigint

		select {
		case <-done:
			return
		default:
			close(done)
		}

	}()

	// Collect metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(fmt.Sprintf("%s:%s", host, profilerPort), nil); err != http.ErrServerClosed {
			log.Error().Err(err).Msgf("Metric server ListenAndServe")
			select {
			case <-done:
				return
			default:
				close(done)
			}
		}
	}()

	// Listen requests
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error().Err(err).Msgf("Server ListenAndServe")
			select {
			case <-done:
				return
			default:
				close(done)
			}
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		log.Error().Err(err).Msgf("Could not shutdown server")
	}
}

func GetAllTask(w http.ResponseWriter, r *http.Request) {
	data := []proto.Task{
		{
			Id:       0,
			Value:    "Task 1",
			Comments: nil,
		},
		{
			Id:       1,
			Value:    "Task 2",
			Comments: nil,
		},
	}
	files := []string{"./assets/html/tasks.gohtml"}
	if len(files) > 0 {
		name := path.Base(files[0])
		tmpl, err := template.New(name).ParseFiles(files...)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to prepare template")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		}
		w.WriteHeader(200)
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to execute template")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}

func GetTask(w http.ResponseWriter, r *http.Request) {
	data := proto.Task{
		Id:         0,
		Value:      "Task 1",
		IsResolved: true,
		Comments:   nil,
	}
	files := []string{"./assets/html/task.gohtml"}
	if len(files) > 0 {
		name := path.Base(files[0])
		tmpl, err := template.New(name).ParseFiles(files...)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to prepare template")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		}
		w.WriteHeader(200)
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to execute template")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}

func InsertTask(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(gitHubUrl + "/gists")
	if err != nil {
		log.Error().Err(err).Msg("Failed to send get request")
		w.WriteHeader(http.StatusInternalServerError)
	}

	switch resp.StatusCode {
	case 200:
		// Parse body
		w.WriteHeader(http.StatusOK)
		log.Debug().Msgf("Response is %v", resp)
	case 422:
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func RemoveTask(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(gitHubUrl + "/gists")
	if err != nil {
		log.Error().Err(err).Msg("Failed to send get request")
		w.WriteHeader(http.StatusInternalServerError)
	}

	switch resp.StatusCode {
	case 200:
		// Parse body
		w.WriteHeader(http.StatusOK)
		log.Debug().Msgf("Response is %v", resp)
	case 422:
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {

}

func Authorize(w http.ResponseWriter, r *http.Request) {

}

func OptionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET, POST, PUT, DELETE")

}
