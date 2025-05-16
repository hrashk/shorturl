package app

var config = struct {
	listenAddr      string
	redirectBaseURL string
}{
	listenAddr:      ":8080",
	redirectBaseURL: "http://localhost:8080",
}
