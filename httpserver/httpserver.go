package httpserver

import (
	"RussianFedoraBot/db"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Server is a main object
type Server struct {
	Addr string
}

const (
	header = `<html>
    <head>
		<style type="text/css">
			TH {
		    	background: #a52a2a; /* Цвет фона */
		    	color: white; /* Цвет текста */
		   	}
		   	TR.even {
    			background: #fff8dc;
   			}
			P.reply {
				color: grey;
			}
		</style>
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
		}
	} else if len(lpath) == 3 {
		if lpath[1] != "" {
			iID, err := strconv.ParseInt(lpath[1], 10, 64)
			if err != nil {
				log.Printf(err.Error())
				w.WriteHeader(http.StatusBadRequest)
				w.Write(parseTemplate("Bad request"))
				return
			}
			log.Printf("Group ID: %d", iID)
			w.Write(parseTemplate(getMessages(iID)))
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
		chatName := chat.UserName
		if chat.UserName == "" {
			chatName = strings.TrimSpace(fmt.Sprintf("%s %s", chat.FirstName, chat.LastName))
		}
		if chat.UserName != "" && (chat.FirstName != "" || chat.LastName != "") {
			names := strings.TrimSpace(chat.FirstName + " " + chat.LastName)
			chatName += fmt.Sprintf(" (%s)", names)
		}
		body += fmt.Sprintf("<li><a href=\"/%d/\">%s</a></li>", chat.ID, chatName)
	}
	body += "</ul>"

	return
}

func getDate(id int64) (body string) {
	// TODO: create it
	return
}

func getMessages(chatID int64) (body string) {
	body += "<table border=0><caption>Messages list</caption>"

	msgs, err := db.GetMessages(chatID)
	if err != nil {
		log.Printf("Error in getMessages: %s", err)
		return ""
	}

	for index, msg := range msgs {
		//body += fmt.Sprintf("<li><a href=\"/%d/\">%s (%s %s)</a></li>", chat.ID, chat.UserName, chat.FirstName, chat.LastName)
		t := time.Unix(int64(msg.Date), 0)
		name := msg.From.UserName
		if msg.From.UserName == "" {
			name = fmt.Sprintf("%s %s", msg.From.FirstName, msg.From.LastName)
		}
		if msg.From.FirstName != "" || msg.From.LastName != "" {
			names := strings.TrimSpace(msg.From.FirstName + " " + msg.From.LastName)
			name += fmt.Sprintf(" (%s)", names)
		}

		msgText := msg.Text
		re := regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
		msgText = re.ReplaceAllString(msgText, `<a href="$0">$0</a>`)

		if msg.ReplyToMessage != nil {
			msgText = fmt.Sprintf("<p class=\"reply\"> > %s</p>\n<p>%s</p>", msg.ReplyToMessage.Text, msgText)
		}

		class := ""
		if index%2 == 0 {
			class = "class=\"even\""
		}

		body += fmt.Sprintf(`
			<tr %s>
				<td class="la" width='12%%'>%s</td>
				<td class="la" width='17%%'><strong>%s</strong></td>
				<td class="la" >%s</td>
			</tr>`, class, t.String(), name, msgText)
	}
	body += "</table>"

	return
}
