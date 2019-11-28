package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/psu/todo-service/pkg/crypt"
	"github.com/psu/todo-service/proto"
	uuid "github.com/satori/go.uuid"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

const (
	selectAllTasksQuery = "SELECT uuid, value, is_resolved FROM public.tasks WHERE author_uuid = $1"
	selectTaskQuery     = "SELECT value, is_resolved FROM public.tasks WHERE author_uuid = $1 AND uuid = $2"
	insertTaskQuery     = "INSERT INTO public.tasks(uuid, value, author_uuid, is_resolved) VALUES ($1, $2, $3, $4)"
	updateTaskQuery     = "UPDATE public.tasks SET is_resolved = $3 WHERE uuid = $1 AND author_uuid = $2"
	deleteTaskQuery     = "DELETE FROM public.tasks WHERE uuid = $1 AND uuid = $2"
	selectUserQuery     = "SELECT uuid, password, salt FROM public.users WHERE login = $1"
	insertUserQuery     = "INSERT INTO public.users(uuid, login, password, salt) VALUES ($1, $2, $3, $4)"
)

var db *sql.DB

// Init database
func InitDatabaseConnection(host string, port string, user string, password string, name string) (err error) {
	// Open connection
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("could not open database connection: %v", err)
	}
	// Test connection
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("could not connect to database: %v", err)
	}
	return
}

// Close db connection
func CloseConnection() (err error) {
	return db.Close()
}

// Select all tasks from database
func SelectAllTasks(userUUID string) (tasks []proto.Task, err error) {
	// Initialize
	selectAllTask, err := db.Prepare(selectAllTasksQuery)
	if err != nil {
		return nil, fmt.Errorf("could not prepare select all query: %v task", err)
	}
	defer func() {
		err = selectAllTask.Close()
		if err != nil {
			log.Error().Msgf("Could not close database connection: %v", err)
		}
	}()
	task := proto.Task{Author: userUUID}
	// Execute query
	rows, err := selectAllTask.Query(userUUID)
	if err != nil {
		return nil, fmt.Errorf("could not select all tasks: %v", err)
	}

	// Fill collection
	for rows.Next() {
		err = rows.Scan(&task.UUID, &task.Value, &task.IsResolved)
		tasks = append(tasks, task)
		if err != nil {
			return nil, fmt.Errorf("could not read query: %v", err)
		}
	}
	return tasks, nil
}

// Select task
func SelectTask(userUUID string, taskUUID string) (task proto.Task, err error) {
	// Initialize
	selectTask, err := db.Prepare(selectTaskQuery)
	if err != nil {
		return proto.Task{}, fmt.Errorf("could not prepare select task query: %v", err)
	}
	defer func() {
		err = selectTask.Close()
		if err != nil {
			log.Error().Msgf("Could not close database connection: %v", err)
		}
	}()

	task.UUID = taskUUID
	task.Author = userUUID

	// Select task
	err = selectTask.QueryRow(task.Author, task.UUID).Scan(
		&task.Value,
		&task.IsResolved,
	)

	if err != nil {
		return proto.Task{}, fmt.Errorf("could not select task: %v", err)
	}
	task.Comments = proto.GenerateComments(task.UUID)

	return task, nil
}

// Insert new task into database
func InsertTask(author string, value string, isResolved bool) (err error) {
	insertTask, err := db.Prepare(insertTaskQuery)
	if err != nil {
		return fmt.Errorf("could not prepare insert query: %v", err)
	}
	defer func() {
		err = insertTask.Close()
		if err != nil {
			log.Error().Err(err).Msgf("Could not close database connection")
		}
	}()
	id := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("could not generate uuid: %v", err)
	}
	_, err = insertTask.Exec(id, author, value, isResolved)
	if err != nil {
		return fmt.Errorf("could not insert task into database: %v", err)
	}
	log.Info().Msgf("Task with uuid = %s is added in database", id)
	return nil
}

// Update task status
func UpdateTask(taskId string, authorId string, isResolved bool) (err error) {
	updateTask, err := db.Prepare(updateTaskQuery)
	if err != nil {
		return fmt.Errorf("could not prepare update query: %v", err)
	}
	defer func() {
		err = updateTask.Close()
		if err != nil {
			log.Error().Err(err).Msgf("Could not close database connection")
		}
	}()
	_, err = updateTask.Exec(taskId, authorId, isResolved)
	if err != nil {
		return fmt.Errorf("could not update task in database: %v", err)
	}
	log.Info().Msgf("Task with uuid = %s is updated in database", taskId)
	return nil
}

// Delete task from database
func DeleteTask(taskId string, userId string) (err error) {
	deleteTask, err := db.Prepare(deleteTaskQuery)
	if err != nil {
		return fmt.Errorf("could not prepare delete query: %v", err)
	}
	defer func() {
		err = deleteTask.Close()
		if err != nil {
			log.Error().Err(err).Msg("Could not close database connection:")
		}
	}()
	_, err = deleteTask.Exec(taskId, userId)
	if err != nil {
		return fmt.Errorf("could not delete task: %v", err)
	}
	log.Info().Msgf("Task with taskId = %s has been deleted", taskId)
	return nil
}

// Sign up user
func SignUp(login string, password string) (string, error) {
	insertUser, err := db.Prepare(insertUserQuery)
	if err != nil {
		return "", fmt.Errorf("could not prepare insert query: %v", err)
	}
	defer func() {
		err = insertUser.Close()
		if err != nil {
			log.Error().Err(err).Msgf("Could not close database connection:")
		}
	}()
	id := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("could not generate uuid: %v", err)
	}
	salt := uuid.NewV4()
	if err != nil {
		return "", fmt.Errorf("could not generate salt: %v", err)
	}
	hashedPassword := crypt.Encrypt([]byte(password), salt.String())

	_, err = insertUser.Exec(id, login, hashedPassword, salt)
	if err != nil {
		return "", fmt.Errorf("could not insert task into database: %v", err)
	}
	log.Info().Msgf("Task with uuid = %s is added in database", id)
	return id.String(), nil
}

// Sign in user
func SignIn(login string, password string) (string, error) {
	// Initialize
	selectUser, err := db.Prepare(selectUserQuery)
	if err != nil {
		return "", fmt.Errorf("could not prepare select user query: %v", err)
	}
	defer func() {
		err = selectUser.Close()
		if err != nil {
			log.Error().Msgf("Could not close database connection: %v", err)
		}
	}()
	user := proto.User{
		Login: login,
	}
	// Execute query
	rows, err := selectUser.Query(user.Login)
	if err != nil {
		return "", fmt.Errorf("could not select user: %v", err)
	}

	// Scan user
	for rows.Next() {
		err = rows.Scan(&user.UUID, &user.Password, &user.Salt)
		if err != nil {
			return "", fmt.Errorf("could not read query: %v", err)
		}
	}

	// Check password
	pass := crypt.Decrypt([]byte(user.Password), user.Salt)
	if string(pass) == user.Password {
		return "", errors.New("invalid password")
	}

	return user.UUID, nil
}
