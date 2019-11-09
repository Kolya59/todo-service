package proto

type Task struct {
	UUID       string    `json:"uuid"`
	Author     string    `json:"author"`
	Value      string    `json:"value"`
	IsResolved bool      `json:"is_resolved"`
	Comments   []Comment `json:"comments"`
}
