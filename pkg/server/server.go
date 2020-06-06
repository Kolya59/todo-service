package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"github.com/Kolya59/todo-service/models"
	"github.com/Kolya59/todo-service/pkg/postgres"
)

const (
	cookieDuration = 30 * time.Minute
)

var (
	tasksUrl string
	loginUrl string
)

func StartServer(host string, port string, profilerPort string) {
	// Create router
	r := chi.NewRouter()
	// Setup routes
	r.Options("/", optionsHandler)

	r.Post("/auth/signin", authorize)
	r.Post("/auth/signup", register)

	r.Get("/tasks", getAllTask)
	r.Post("/tasks", insertTask)

	r.Get("/tasks/{id}", getTask)
	r.Put("/tasks/{id}", updateTaskStatus)
	r.Delete("/tasks/{id}", removeTask)

	// File routes
	r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./assets/html/auth.gohtml")
	})

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

	loginUrl = fmt.Sprintf("%v:%v/auth", host, port)
	tasksUrl = fmt.Sprintf("%v:%v/tasks", host, port)

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

func auth(r *http.Request) (string, error) {
	c, err := r.Cookie("id")
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func getAllTask(w http.ResponseWriter, r *http.Request) {
	id, err := auth(r)
	if err != nil || id == "" {
		log.Info().Err(err).Msg("Failed to authorize user")
		http.Redirect(w, r, loginUrl, http.StatusUnauthorized)
		return
	}

	tasks, err := postgres.SelectAllTasks(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tasks")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	files := []string{"./assets/html/tasks.gohtml"}
	if len(files) > 0 {
		name := path.Base(files[0])
		tmpl, err := template.New(name).ParseFiles(files...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to prepare template")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(200)
		err = tmpl.Execute(w, tasks)
		if err != nil {
			log.Error().Err(err).Msg("Failed to execute template")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}

func getTask(w http.ResponseWriter, r *http.Request) {
	userId, err := auth(r)
	if err != nil || userId == "" {
		log.Info().Err(err).Msg("Failed to authorize user")
		http.Redirect(w, r, loginUrl, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	task, err := postgres.SelectTask(userId, id)
	files := []string{"./assets/html/task.gohtml"}
	if len(files) > 0 {
		name := path.Base(files[0])
		tmpl, err := template.New(name).ParseFiles(files...)
		if err != nil {
			log.Error().Err(err).Msg("Failed to prepare template")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		}
		w.WriteHeader(200)
		err = tmpl.Execute(w, task)
		if err != nil {
			log.Error().Err(err).Msg("Failed to execute template")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}

func insertTask(w http.ResponseWriter, r *http.Request) {
	task := &models.Task{}
	var err error

	task.Author, err = auth(r)
	if err != nil || task.Author == "" {
		log.Info().Err(err).Msg("Failed to authorize user")
		http.Redirect(w, r, loginUrl, http.StatusUnauthorized)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(data, task)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshall body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := postgres.InsertTask(task.Value, task.Author, false)
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprint("Failed to insert task")))
		return
	}
	response, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprint("Failed to insert task")))
		return
	}
	_, err = w.Write(response)
	w.WriteHeader(200)
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprint("Failed to insert task")))
		return
	}
}

func removeTask(w http.ResponseWriter, r *http.Request) {
	userId, err := auth(r)
	if err != nil || userId == "" {
		log.Info().Err(err).Msg("Failed to authorize user")
		http.Redirect(w, r, loginUrl, http.StatusUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	err = postgres.DeleteTask(userId, id)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to delete task %v", id)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func updateTaskStatus(w http.ResponseWriter, r *http.Request) {
	userId, err := auth(r)
	if err != nil || userId == "" {
		log.Info().Err(err).Msg("Failed to authorize user")
		http.Redirect(w, r, loginUrl, http.StatusUnauthorized)
		return
	}

	type requestBody struct {
		IsResolved bool `json:"is_resolved"`
	}
	request := &requestBody{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(data, request)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	id := chi.URLParam(r, "id")
	err = postgres.UpdateTask(id, userId, request.IsResolved)
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprint("Failed to update task")))
		return
	}
	w.WriteHeader(200)
}

func authorize(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode body")
	}
	err = json.Unmarshal(data, user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshall body")
	}

	id, err := postgres.SignIn(user.Login, user.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to sign in")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "id",
		Value:    id,
		Expires:  time.Now().Add(cookieDuration),
		Secure:   false,
		HttpOnly: true,
		Path:     "/",
	})

	http.Redirect(w, r, tasksUrl, http.StatusOK)
}

func register(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode body")
	}
	err = json.Unmarshal(data, user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshall body")
	}

	id, err := postgres.SignUp(user.Login, user.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to sign up")
		if err.Error() == "user is exist" {
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "id",
		Value:    id,
		Expires:  time.Now().Add(cookieDuration),
		Secure:   false,
		HttpOnly: true,
		Path:     "/",
	})

	http.Redirect(w, r, tasksUrl, http.StatusOK)
}

func optionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET, POST, PUT, DELETE")

}
