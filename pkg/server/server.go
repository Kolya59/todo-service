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
	"github.com/psu/todo-service/pkg/postgres"
	"github.com/psu/todo-service/proto"
	"github.com/rs/zerolog/log"
)

func StartServer(host string, port string, profilerPort string) {
	// Create router
	r := chi.NewRouter()
	// Setup routes
	r.Options("/", OptionsHandler)

	r.Post("/auth/signin", Authorize)
	r.Post("/auth/signup", Register)

	r.Get("/tasks", GetAllTask)
	r.Post("/tasks", InsertTask)

	r.Get("/tasks/{id}", GetTask)
	r.Put("/task/{id}", UpdateTaskStatus)
	r.Delete("/tasks/{id}", RemoveTask)

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
	type request struct {
		Id string `json:"id"`
	}
	userId := request{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(data, userId)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshall body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tasks, err := postgres.SelectAllTasks(userId.Id)
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

func GetTask(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		UserId string `json:"user_id"`
		TaskId string `json:"task_id"`
	}
	request := requestBody{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(data, request)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshall body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	task, err := postgres.SelectTask(request.UserId, request.TaskId)
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
		err = tmpl.Execute(w, task)
		if err != nil {
			log.Panic().Err(err).Msg("Failed to execute template")
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}

func InsertTask(w http.ResponseWriter, r *http.Request) {
	task := proto.Task{}
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
	err = postgres.InsertTask(task.Author, task.Value, task.IsResolved)
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprint("Failed to insert task")))
		return
	}
	response, err := json.Marshal(task)
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

func RemoveTask(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		UserId string `json:"user_id"`
		TaskId string `json:"task_id"`
	}
	request := requestBody{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(data, request)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshall body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = postgres.DeleteTask(request.UserId, request.TaskId)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to delete task %v", request.TaskId)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func UpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		UserId     string `json:"user_id"`
		TaskId     string `json:"task_id"`
		IsResolved bool   `json:"is_resolved"`
	}
	request := requestBody{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to decode body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(data, request)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshall body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = postgres.UpdateTask(request.TaskId, request.UserId, request.IsResolved)
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprint("Failed to update task")))
		return
	}
	w.WriteHeader(200)
}

func Authorize(w http.ResponseWriter, r *http.Request) {
	user := proto.User{}
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
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(id))
}

func Register(w http.ResponseWriter, r *http.Request) {
	user := proto.User{}
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
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(id))
}

func OptionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET, POST, PUT, DELETE")

}
