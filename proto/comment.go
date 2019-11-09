package proto

import (
	"math/rand"

	"github.com/google/uuid"
)

type Comment struct {
	UUID   string `json:"uuid"`
	Value  string `json:"value"`
	Author string `json:"author"`
	TaskId string `json:"task_id"`
}

func GenerateComments(taskID string) []Comment {
	n := rand.Intn(10)
	comments := make([]Comment, n)
	for i := 0; i < n; i++ {
		comments = append(comments, Comment{
			UUID:   uuid.New().String(),
			Value:  "Best comment",
			Author: "Implement it",
			TaskId: taskID,
		})
	}
	return comments
}
