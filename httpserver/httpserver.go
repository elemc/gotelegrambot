package httpserver

import (
	"RussianFedoraBot/db"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Server is a main object
type Server struct {
	Addr string
}

const (
	header = `<html>
    <head>
    </head>

    <body>`
	footer = `</body>
    </html>`
)

// Start method starts http server
func (s *Server) Start() {
	http.HandleFunc("/", s.handlerIndex)
	http.ListenAndServe(s.Addr, nil)
}

func (s *Server) handlerIndex(w http.ResponseWriter, r *http.Request) {
	lpath := strings.Split(r.URL.Path, "/")
	for index, path := range lpath {
		log.Printf("Path [%d]: %s", index, path)
	}

	w.WriteHeader(http.StatusOK)
	if len(lpath) == 2 {
		if lpath[0] == "" && lpath[1] == "" {
			w.Write(parseTemplate(getMain()))
		} else if lpath[1] != "" {
			iID, err := strconv.ParseInt(lpath[1], 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write(parseTemplate("Bad request"))
				return
			}
			w.Write(parseTemplate(getDate(iID)))
		}
	}
}

func parseTemplate(body string) []byte {
	res := fmt.Sprintf("%s\n%s\n%s", header, body, footer)
	return []byte(res)
}

func getMain() (body string) {
	body += "<h1>Chat list</h1><ul>"

	chats, err := db.GetChats()
	if err != nil {
		log.Printf("Error in getMain: %s", err)
		return ""
	}

	for _, chat := range chats {
		body += fmt.Sprintf("<li><a href=\"/%d/\">%s (%s %s)</a></li>", chat.ID, chat.UserName, chat.FirstName, chat.LastName)
	}

	return
}

func getDate(id int64) (body string) {
	// TODO: create it
	return
}

func getMessages(id int64) {

}
