package proto

type User struct {
	UUID     string `json:"uuid"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
}
