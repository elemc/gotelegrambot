package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elemc/gotelegrambot/db"

	"github.com/gin-gonic/gin"
	"gopkg.in/telegram-bot-api.v4"
)

// Server is a main object
type Server struct {
	Addr          string
	Bot           *tgbotapi.BotAPI
	PhotoCache    PhotosCache
	FileCache     FilesCache
	APIKey        string
	CensList      []string
	StaticDirPath string
}

const (
	header = `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN"
        "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
	<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
    <head>
		<title>Telegram logs</title>
		<meta charset="utf-8;" />
		<style type="text/css">
			TH {
		    	background: #FFFFFF; /* Цвет фона */
		    	color: white; /* Цвет текста */
		   	}
			TD {
				vertical-align: top;
			}
		   	TR.even {
    			background: #F0F4F7;
   			}
			P.reply {
				color: grey;
			}
		</style>
    </head>
    <body>
	<h2><a href="/">Telegram logs</a></h2>`
	footer = `</body>
</html>`
	tableBegin = `<table border="0"><caption>%s</caption>`
	classEven  = `class="even"`
	tableEnd   = `</table>`
)

// Start method starts http server
func (s *Server) Start() {
	s.UpdatePhotoCache()
	go s.updatePhotoCacheServer()

	r := gin.Default()

	r.StaticFS("/static", http.Dir(s.StaticDirPath))
	r.GET("/chat/:chat_id/:year/:month/:day", s.dayPage)
	r.GET("/chat/:chat_id/:year/:month", s.monthPage)
	r.GET("/chat/:chat_id/:year", s.yearPage)
	r.GET("/chat/:chat_id/", s.chatPage)

	r.GET("/", s.mainPage)

	r.Run(s.Addr)
}

func (s *Server) mainPage(c *gin.Context) {
	page := parseTemplate(s.getMain())
	c.Data(http.StatusOK, "text/html", page)
}

func (s *Server) chatPage(c *gin.Context) {
	strChatID := c.Param("chat_id")
	chatID, err := strconv.ParseInt(strChatID, 10, 64)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}

	page := parseTemplate(s.getYears(chatID))
	// page := parseTemplate(s.getMessages(chatID))
	c.Data(http.StatusOK, "text/html", page)
}

func (s *Server) yearPage(c *gin.Context) {
	strChatID := c.Param("chat_id")
	strYear := c.Param("year")
	chatID, err := strconv.ParseInt(strChatID, 10, 64)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	year, err := strconv.Atoi(strYear)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}

	page := parseTemplate(s.getMonths(chatID, year))
	// page := parseTemplate(s.getMessages(chatID))
	c.Data(http.StatusOK, "text/html", page)
}

func (s *Server) monthPage(c *gin.Context) {
	strChatID := c.Param("chat_id")
	strYear := c.Param("year")
	strMonth := c.Param("month")
	chatID, err := strconv.ParseInt(strChatID, 10, 64)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	year, err := strconv.Atoi(strYear)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	month, err := strconv.Atoi(strMonth)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}

	page := parseTemplate(s.getDates(chatID, year, month))
	c.Data(http.StatusOK, "text/html", page)
}

func (s *Server) dayPage(c *gin.Context) {
	strChatID := c.Param("chat_id")
	strYear := c.Param("year")
	strMonth := c.Param("month")
	strDay := c.Param("day")
	chatID, err := strconv.ParseInt(strChatID, 10, 64)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	year, err := strconv.Atoi(strYear)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	month, err := strconv.Atoi(strMonth)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}
	day, err := strconv.Atoi(strDay)
	if err != nil {
		c.String(http.StatusOK, err.Error())
		return
	}

	beginTime := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
	endTime := time.Date(year, time.Month(month), day, 23, 59, 59, 100, time.Local)

	page := parseTemplate(s.getMessages(chatID, beginTime, endTime))
	c.Data(http.StatusOK, "text/html", page)
}

func (s *Server) updatePhotoCacheServer() {
	for {
		time.Sleep(time.Minute * 5)
		log.Printf("Update phtoto cache started...")
		s.UpdatePhotoCache()
		log.Printf("Update cens database started...")
		s.FillCens()
	}
}

func parseTemplate(body string) []byte {
	result := fmt.Sprintf("%s\n%s\n%s", header, body, footer)
	return []byte(result)
}

func (s *Server) getMain() (body string) {
	body += fmt.Sprintf(tableBegin, "Chats")

	chats, err := db.GetChats()
	if err != nil {
		log.Printf("Error in getMain: %s", err)
		return ""
	}

	for index, chat := range chats {
		chatName := chat.Title
		if chat.Title == "" {
			chatName = chat.UserName
		}

		if chatName == "" {
			chatName = strings.TrimSpace(fmt.Sprintf("%s %s", chat.FirstName, chat.LastName))
		}
		if chatName != "" && (chat.FirstName != "" || chat.LastName != "") {
			names := strings.TrimSpace(chat.FirstName + " " + chat.LastName)
			chatName += fmt.Sprintf(" (%s)", names)
		}

		class := ""
		if index%2 == 0 {
			class = classEven
		}

		body += fmt.Sprintf(`
			<tr %s>
				<td class="la"><a href="/chat/%d/">%s</a></td>
			</tr>`, class, chat.ID, chatName)

	}
	body += tableEnd

	return
}

func getDate(id int64) (body string) {
	// TODO: create it
	return
}

func (s *Server) getMessages(chatID int64, beginTime, endTime time.Time) (body string) {
	body += fmt.Sprintf(tableBegin, "Messages")

	msgs, err := db.GetMessagesByDate(chatID, beginTime, endTime)
	if err != nil {
		log.Printf("Error in getMessages: %s", err)
		return ""
	}

	for index, msg := range msgs {
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
			lt := time.Unix(int64(msg.ReplyToMessage.Date), 0)
			replyLink := fmt.Sprintf("/chat/%d/%d/%d/%d#%s", msg.Chat.ID, lt.Year(), lt.Month(), lt.Day(), lt.Format("15:04:05"))
			msgText = fmt.Sprintf(`<p class="reply"> <a href="%s">></a> %s</p><p>%s</p>`, replyLink, msg.ReplyToMessage.Text, url.QueryEscape(msgText))
		}

		class := ""
		if index%2 == 0 {
			class = classEven
		}

		photo := s.GetPhotoFileName(int64(msg.From.ID))
		timeStr := t.Format("15:04:05")

		if msg.Audio != nil {
			msgText += fmt.Sprintf(`<p><a href="/%s">Audio in message</a></p>`, s.GetFileNameByFileID(msg.Chat.ID, msg.Audio.FileID))
		}
		if msg.Document != nil {
			msgText += fmt.Sprintf(`<p><a href="/%s">Document in message</a></p>`, s.GetFileNameByFileID(msg.Chat.ID, msg.Document.FileID))
		}
		if msg.Photo != nil {
			msgText += "<p>"
			f := (*msg.Photo)[len(*msg.Photo)-1]
			//for _, f := range *msg.Photo {
			photoName := s.GetFileNameByFileIDURL(msg.Chat.ID, f.FileID)
			msgText += fmt.Sprintf(`<p><a href="/%s"><img src="/%s"></img></a>`, photoName, photoName)
			//}
			msgText += "</p>"
		}
		if msg.Sticker != nil {
			msgText += fmt.Sprintf(`<p><img src="/%s"></img></p>`, s.GetFileNameByFileIDURL(msg.Chat.ID, msg.Sticker.FileID))
		}
		if msg.Video != nil {
			msgText += fmt.Sprintf(`<p><a href="/%s">Video in message</a></p>`, s.GetFileNameByFileIDURL(msg.Chat.ID, msg.Video.FileID))
		}
		if msg.Voice != nil {
			msgText += fmt.Sprintf(`<p><a href="/%s">Voice in message</a></p>`, s.GetFileNameByFileIDURL(msg.Chat.ID, msg.Voice.FileID))
		}

		body += fmt.Sprintf(`
			<tr %s>
				<td class="la" align="center" width='3%%'><img src="/%s" height="30px" width="30px"></img></td>
				<td class="la" align="center" width='5%%'><a id="%s" name="%s" href="#%s" class="time">%s</td>
				<td class="la" width='17%%'><strong>%s</strong></td>
				<td class="la">%s</td>
				<td style="display:none;">%d</td>
			</tr>`, class, photo, timeStr, timeStr, timeStr, timeStr, name, url.QueryEscape(msgText), msg.MessageID)
	}
	body += tableEnd

	return
}

func (s *Server) getYears(chatID int64) (body string) {
	body += fmt.Sprintf(tableBegin, "Years")

	dates, err := db.GetYears(chatID)
	if err != nil {
		log.Printf("Error in GetYears for chat %d: %s", chatID, err)
		return ""
	}
	for index, date := range dates {
		class := ""
		if index%2 == 0 {
			class = classEven
		}
		body += fmt.Sprintf(`
			<tr %s>
				<td class="la"><a href="/chat/%d/%s">%s</a></td>
			</tr>`, class, chatID, date, date)

	}
	body += tableEnd

	return
}

func (s *Server) getMonths(chatID int64, year int) (body string) {
	body += fmt.Sprintf(tableBegin, "Months")

	dates, err := db.GetMonthList(chatID, year)
	if err != nil {
		log.Printf("Error in GetYears for chat %d: %s", chatID, err)
		return ""
	}
	for index, date := range dates {
		class := ""
		if index%2 == 0 {
			class = classEven
		}
		body += fmt.Sprintf(`
			<tr %s>
				<td class="la" ><a href="/chat/%d/%d/%d">%s</a></td>
			</tr>`, class, chatID, year, date, date.String())

	}
	body += tableEnd

	return
}

func (s *Server) getDates(chatID int64, year int, month int) (body string) {
	body += fmt.Sprintf(tableBegin, "Dates")

	dates, err := db.GetDates(chatID, year, month)
	if err != nil {
		log.Printf("Error in GetYears for chat %d: %s", chatID, err)
		return ""
	}
	for index, date := range dates {
		class := ""
		if index%2 == 0 {
			class = classEven
		}
		body += fmt.Sprintf(`
			<tr %s>
				<td class="la" ><a href="/chat/%d/%d/%d/%d">%02d</a></td>
			</tr>`, class, chatID, year, month, date, date)

	}
	body += tableEnd

	return
}
