package proto

type Task struct {
	Id         int       `json:"id"`
	Value      string    `json:"value"`
	IsResolved bool      `json:"is_resolved"`
	Comments   []Comment `json:"comments"`
}

type Comment struct {
	Id     int    `json:"id"`
	Value  string `json:"value"`
	Author string `json:"author"`
	TaskId int    `json:"task_id"`
}
