package handle

import "net/http"

type HandleHTTP struct {
	Path   string
	Handle func(http.ResponseWriter, *http.Request)
}
