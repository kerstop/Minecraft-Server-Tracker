// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package servers

type Server struct {
	ID   int64  `json:"id"`
	Url  string `json:"url"`
	Port int64  `json:"port"`
}

type ServerStatus struct {
	ServerID int64 `json:"server_id"`
	Time     int64 `json:"time"`
	Count    int64 `json:"count"`
}