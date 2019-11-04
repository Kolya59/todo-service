package proto

type Task struct {
	UUID       string    `json:"uuid"`
	Value      string    `json:"value"`
	IsResolved bool      `json:"is_resolved"`
	Comments   []Comment `json:"comments"`
}

type Comment struct {
	UUID   string `json:"uuid"`
	Value  string `json:"value"`
	Author string `json:"author"`
	TaskId int    `json:"task_id"`
}
