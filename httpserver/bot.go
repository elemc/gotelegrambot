package httpserver

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/elemc/gotelegrambot/db"

	"gopkg.in/telegram-bot-api.v4"
)

// PhotosCache type for store users photo filenames by id
type PhotosCache map[int64]string

// FilesCache type for store files
type FilesCache map[string]string

// UpdatePhotoCache function update photos cache of users
func (s *Server) UpdatePhotoCache() {
	users, err := db.GetUsers()
	if err != nil {
		log.Printf("Error in UpdatePhotoCache: %s", err)
		return
	}

	s.PhotoCache = make(PhotosCache) // new cache

	for _, user := range users {
		s.GetPhoto(int64(user.ID))
	}
}

// GetPhotoFileName returns name photo file
func (s *Server) GetPhotoFileName(userID int64) (result string) {
	if fn, ok := s.PhotoCache[userID]; ok {
		result = getFileName("static", fn)
	} else {
		result = getFileName("static", "nobody.png")
	}
	return
}

// GetFileNameByFileID returns file name by index
func (s *Server) GetFileNameByFileID(chatID int64, fileID string) (filename string) {
	f, err := db.GetFile(fileID, chatID)
	if err != nil {
		// try to download it
		s.GetFile(fileID, chatID)
		f, err = db.GetFile(fileID, chatID)
		if err != nil {
			log.Printf("Error in GetFileNameByFileID with FileID [%s]: %s", fileID, err)
			return "missing-data"
		}
	}
	filename = filepath.Join(s.StaticDirPath, f.FilePath)

	return
}

// GetFileNameByFileIDURL returns file name by index
func (s *Server) GetFileNameByFileIDURL(chatID int64, fileID string) (filename string) {
	f, err := db.GetFile(fileID, chatID)
	if err != nil {
		// try to download it
		s.GetFile(fileID, chatID)
		f, err = db.GetFile(fileID, chatID)
		if err != nil {
			log.Printf("Error in GetFileNameByFileID with FileID [%s]: %s", fileID, err)
			return "missing-data"
		}
	}
	filename = filepath.Join("static/", f.FilePath)

	return
}

// GetPhoto function download user photo and return file name for html tag img
func (s *Server) GetPhoto(chatID int64) {
	config := tgbotapi.NewUserProfilePhotos(int(chatID))
	photos, err := s.Bot.GetUserProfilePhotos(config)
	if err != nil {
		if err.Error() == "Bad Request: user not found" {
			return
		}
		log.Printf("Error in GetPhoto for ID %d: %s", chatID, err.Error())
		return
	}
	if photos.TotalCount == 0 {
		return
	}
	res := photos.Photos[0][0]

	link, err := s.Bot.GetFileDirectURL(res.FileID)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	filename := fmt.Sprintf("%d.jpg", chatID)
	err = downloadImage(link, getFileName(s.StaticDirPath, filename))
	if err != nil {
		log.Printf("Error in downloadImage: %s", err)
		return
	}
	s.PhotoCache[chatID] = filename

	return
}

// GetFile function for get file from telegram
func (s *Server) GetFile(fileID string, chatID int64) {
	fc := tgbotapi.FileConfig{}
	fc.FileID = fileID
	f, err := s.Bot.GetFile(fc)
	if err != nil {
		log.Printf("Error in GetFile for FileID [%s]: %s", fileID, err)
		return
	}

	log.Printf(f.FilePath)

	// check directory
	dir := filepath.Dir(f.FilePath)
	path := filepath.Join(s.StaticDirPath, dir)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		log.Printf("Error in MkdirAll for FileID [%s]: %s", fileID, err)
		return
	}

	filename := filepath.Join(s.StaticDirPath, f.FilePath)
	err = downloadImage(f.Link(s.APIKey), filename)
	if err != nil {
		log.Printf("Error in MkdirAll for FileID [%s]: %s", fileID, err)
		return
	}
	//s.FileCache[f.FileID] = filepath.Join("static", f.FilePath)
	err = db.SaveFile(&f, chatID)
	if err != nil {
		log.Printf("Error in SaveFile for FileID [%s]: %s", fileID, err)
	}
}

// SendMessage function send message to given user
// msgText - is the message text
// chatID - ID for chat (user id or chat id)
// user - User struct database store
// replyID - id messages for reply or 0
func (s *Server) SendMessage(msgText string, chatID int64, replyID int) {
	// buttonText := notifEnable
	// if user != nil {
	// 	if user.NotificationEnabled {
	// 		buttonText = notifDisable
	// 	}
	//
	// }

	msg := tgbotapi.NewMessage(chatID, msgText)

	// Keyboard
	// k := tgbotapi.NewKeyboardButtonContact("Отправить номер телефона")
	// kn := tgbotapi.NewKeyboardButton(buttonText)
	// rows := tgbotapi.NewKeyboardButtonRow(k, kn)
	// rm := tgbotapi.NewReplyKeyboard(rows)
	// msg.ReplyMarkup = rm

	if replyID != 0 {
		msg.ReplyToMessageID = replyID
	}

	_, err := s.Bot.Send(msg)
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
}

// CommandHandler function for handle commands for bot
func (s *Server) CommandHandler(msg *tgbotapi.Message) {
	if msg == nil {
		return
	}
	switch msg.Command() {
	case "start":
		s.SendMessage("Привет", msg.Chat.ID, msg.MessageID)
	case "help":
		s.SendHelp(msg)
	case "ping":
		s.SendPing(msg)
	case "ban":
		s.BanUnbanUser(msg, true)
	case "unban":
		s.BanUnbanUser(msg, false)
	case "banlist":
		s.BanList(msg)
	case "clearcens":
		s.ClearCens(msg)
	case "mycens":
		s.GetCensLevel(msg)
	case "warn":
		s.WarnAdd(msg)
	case "clearwarn":
		s.WarnClear(msg)
	case "mywarn":
		s.GetWarnLevel(msg)
	default:
		log.Printf("Unknown command: %s", msg.Command())
		// 	if msg.
		// 	s.SendMessage("Неизвестная команда", msg.Chat.ID, msg.MessageID)
	}
}

// SendError simple shell for SendMessage with error
func (s *Server) SendError(msgText string, msg *tgbotapi.Message) {
	s.SendMessage(msgText, msg.Chat.ID, msg.MessageID)
}

// UserIsAdmin returns user is admin or not
func (s *Server) UserIsAdmin(userID int, chat *tgbotapi.Chat) (ok bool, err error) {
	if chat == nil {
		return false, fmt.Errorf("Chat pointer is nil")
	}
	cc := tgbotapi.ChatConfig{}
	if chat.IsSuperGroup() || chat.IsGroup() {
		cc.SuperGroupUsername = "@" + chat.UserName
	} else {
		cc.ChatID = chat.ID
	}
	var admins []tgbotapi.ChatMember
	if admins, err = s.Bot.GetChatAdministrators(cc); err != nil {
		log.Printf("Error in GetChatAdministrators: %s", err)
		return
	}

	ok = false
	for _, admin := range admins {
		if admin.User.ID == userID {
			ok = true
			break
		}
	}
	return
}

// UserIsBanned returns ban status user true or false
func (s *Server) UserIsBanned(userID int, chat *tgbotapi.Chat) (banned bool, err error) {
	cc := tgbotapi.ChatConfigWithUser{}
	if chat.IsSuperGroup() || chat.IsGroup() {
		cc.SuperGroupUsername = "@" + chat.UserName
	} else {
		cc.ChatID = chat.ID
	}
	cc.UserID = userID

	member, err := s.Bot.GetChatMember(cc)
	if err != nil {
		return
	}

	banned = member.WasKicked()
	return
}

// BanList method returns ban list
func (s *Server) BanList(msg *tgbotapi.Message) {
	users, err := db.GetUsers()
	if err != nil {
		log.Printf("Error in GetUsers in BanList: %s", err)
		return
	}

	var bannedList []string
	for _, user := range users {
		banned, err := s.UserIsBanned(user.ID, msg.Chat)
		if err != nil {
			log.Printf("Error in UserIsBanned: %s", err)
			continue
		}
		if banned {
			bannedList = append(bannedList, user.String())
		}
	}

	if len(bannedList) == 0 {
		s.SendMessage("Ура! Мы чисты! Забаненых нет", msg.Chat.ID, msg.MessageID)
		return
	}
	msgText := fmt.Sprintf("Список забанненных лиц:\n%s", strings.Join(bannedList, "\n"))
	s.SendMessage(msgText, msg.Chat.ID, msg.MessageID)
}

// BanUnbanUser method ban selected user
func (s *Server) BanUnbanUser(msg *tgbotapi.Message, ban bool) {
	isAdmin, err := s.UserIsAdmin(msg.From.ID, msg.Chat)
	if err != nil {
		return
	}
	if !isAdmin {
		s.SendError("Не удалось установить Вашу причастность к администраторам группы!", msg)
		return
	}

	user, err := db.GetUser(msg.CommandArguments())
	if err != nil {
		errStrings := strings.Split(err.Error(), "\n")
		if len(errStrings) > 1 {
			switch errStrings[0] {
			case "User not found":
				s.SendError(fmt.Sprintf("Пользователь %s не найден", msg.CommandArguments()), msg)
				return
			case "Many users":
				s.SendError(fmt.Sprintf("Найдено более одного пользователя, уточните:\n%s", strings.Join(errStrings[1:], "\n")), msg)
				return
			default:
				s.SendError(fmt.Sprintf("Произошла неизвестная ошибка при поиске пользователя: %s", err.Error()), msg)
				return
			}
		}
	}

	if user != nil {
		userIsAdmin, _ := s.UserIsAdmin(user.ID, msg.Chat)
		if userIsAdmin {
			s.SendError(fmt.Sprintf("Пользователь [%s] является администратором группы. Администраторов банить нельзя! Они хорошие!", user.String()), msg)
			return
		}
	} else {
		s.SendError(fmt.Sprintf("Пользователь [%s] не найден", msg.CommandArguments()), msg)
		return
	}

	ok, err := s.kickUser(user.ID, msg.Chat, ban)

	if err != nil {
		log.Printf("Error in KickChatMember: %s", err)
		return
	}
	if ok {
		s.SendMessage("Успешно выполнено.", msg.Chat.ID, msg.MessageID)
	}
}

// SendPing sends joke ping to chat
func (s *Server) SendPing(msg *tgbotapi.Message) {
	r := rand.New(rand.NewSource(int64(msg.From.ID)))
	r.Seed(int64(msg.MessageID))

	// Fix issue #1
	if r.Int()%6 == 0 {
		s.SendMessage("Request timed out", msg.Chat.ID, msg.MessageID)
		return
	}
	pingMsg := fmt.Sprintf("%s пинг от тебя %3.3f", msg.From.String(), r.Float32())
	s.SendMessage(pingMsg, msg.Chat.ID, msg.MessageID)
}

// SendHelp sends help message to chat
func (s *Server) SendHelp(msg *tgbotapi.Message) {
	helpMsg :=
		`Помощь по командам бота.
/start - приветствие (стандартная для любого бота Telegram)
/ban @username - забанить пользователя в группе (бот должен иметь административные права в группе)
/unban @username - разбанить пользователя в группе (бот должен иметь административные права в группе)
/banlist - показать список забаненых пользователей
/clearcens - очистить счетчик бранных слов
/mycens - показать собственный счетчик бранных слов
/ping - шуточный пинг`
	s.SendMessage(helpMsg, msg.Chat.ID, msg.MessageID)
}

// FillCens load censore database
func (s *Server) FillCens() {
	f, err := os.Open(filepath.Join(s.StaticDirPath, "mat.txt"))
	if err != nil {
		log.Printf("Error in open mat.txt: %s", err)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("Error in reading mat.txt: %s", err)
		return
	}

	s.CensList = []string{}

	words := strings.Split(string(data), ",")
	for _, word := range words {
		sWord := strings.TrimSpace(word)
		if sWord == "" {
			continue
		}
		s.CensList = append(s.CensList, sWord)
	}
	log.Printf("Cens database filled.")
}

// Cens method for censore messages
func (s *Server) Cens(msg *tgbotapi.Message) {
	lineList := strings.Split(msg.Text, "\n")
	var wordList []string
	for _, line := range lineList {
		wl := strings.Split(line, " ")
		wordList = append(wordList, wl...)
	}
	for _, word := range wordList {
		uWord := strings.ToUpper(word)
		for _, mWord := range s.CensList {
			umWord := strings.ToUpper(mWord)
			if umWord == uWord {
				s.censWord(msg, mWord)
			}
		}
	}
}

// ClearCens command for clean censore level
func (s *Server) ClearCens(msg *tgbotapi.Message) {
	isAdmin, err := s.UserIsAdmin(msg.From.ID, msg.Chat)
	if err != nil {
		return
	}
	if !isAdmin {
		s.SendError("Не удалось установить Вашу причастность к администраторам группы!", msg)
		return
	}

	user, err := db.GetUser(msg.CommandArguments())
	if err != nil {
		errStrings := strings.Split(err.Error(), "\n")
		if len(errStrings) > 1 {
			switch errStrings[0] {
			case "User not found":
				s.SendError(fmt.Sprintf("Пользователь %s не найден", msg.CommandArguments()), msg)
				return
			case "Many users":
				s.SendError(fmt.Sprintf("Найдено более одного пользователя, уточните:\n%s", strings.Join(errStrings[1:], "\n")), msg)
				return
			default:
				s.SendError(fmt.Sprintf("Произошла неизвестная ошибка при поиске пользователя: %s", err.Error()), msg)
				return
			}
		}
	}

	if user == nil {
		s.SendError(fmt.Sprintf("Пользователь %s не найден.", msg.CommandArguments()), msg)
		return
	}

	err = db.ClearCensLevel(user)
	if err != nil {
		log.Printf("Error in ClearCens -> ClearCensLevel: %s", err)
		return
	}
	s.SendError("Выполнено успешно.", msg)
}

// GetCensLevel send message with current censore level for user
func (s *Server) GetCensLevel(msg *tgbotapi.Message) {
	currentLevel, err := db.GetCensLevel(msg.From)
	if err != nil {
		if err.Error() == "Key not found." {
			s.SendError("Ты чист душой!", msg)
			return
		}
		log.Printf("Error in GetCensLevel -> GetCensLevel: %s", err)
		return
	}
	s.SendError(fmt.Sprintf("Твой личный счетчик бранных слов: %d", currentLevel), msg)
}

func (s *Server) censWord(msg *tgbotapi.Message, mWord string) {
	log.Printf("[%s] cens word [%s] in text [%s]", msg.From.String(), mWord, msg.Text)
	s.SendError(fmt.Sprintf("Перестаньте сказать, %s! Вы не на привозе!", msg.From.String()), msg)
	cur, err := db.AddCensLevel(msg.From)
	if err != nil {
		log.Printf("Error in AddCensLevel: %s", err)
		return
	}
	if cur > 5 {
		userIsAdmin, _ := s.UserIsAdmin(msg.From.ID, msg.Chat)
		if userIsAdmin {
			return
		}

		ok, err := s.kickUser(msg.From.ID, msg.Chat, true)

		if err != nil {
			log.Printf("Error in KickChatMember: %s", err)
			return
		}
		if ok {
			s.SendError(fmt.Sprintf("Поздравляю, %s! Вы превысили количество бранных слов в году и выбываете из чата!", msg.From.String()), msg)
		}
	}
	return
}

func getFileName(staticDir, fn string) string {
	return filepath.Join(staticDir, fn)
}

func downloadImage(url string, filename string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return
	}
	return
}

func (s *Server) kickUser(userID int, chat *tgbotapi.Chat, ban bool) (ok bool, err error) {
	ok = false
	config := tgbotapi.ChatMemberConfig{}
	config.UserID = userID
	if chat.IsSuperGroup() || chat.IsGroup() {
		config.SuperGroupUsername = "@" + chat.UserName
		log.Printf("Kick from %s", config.SuperGroupUsername)
	} else {
		config.ChatID = chat.ID
	}

	var resp tgbotapi.APIResponse
	if ban {
		resp, err = s.Bot.KickChatMember(config)
	} else {
		resp, err = s.Bot.UnbanChatMember(config)
	}
	if err != nil {
		return
	}
	if !resp.Ok {
		err = fmt.Errorf("Не удалось забанить/разбанить пользователя: code=%d description: %s", resp.ErrorCode, resp.Description)
	} else {
		ok = true
	}
	return
}

func (s *Server) WarnAdd(msg *tgbotapi.Message) {
	var (
		user *tgbotapi.User
		err  error
	)
	if user, err = db.GetUser(msg.CommandArguments()); err != nil {
		errStrings := strings.Split(err.Error(), "\n")
		if len(errStrings) > 1 {
			switch errStrings[0] {
			case "User not found":
				s.SendError(fmt.Sprintf("Пользователь %s не найден", msg.CommandArguments()), msg)
				return
			case "Many users":
				s.SendError(fmt.Sprintf("Найдено более одного пользователя, уточните:\n%s", strings.Join(errStrings[1:], "\n")), msg)
				return
			default:
				s.SendError(fmt.Sprintf("Произошла неизвестная ошибка при поиске пользователя: %s", err.Error()), msg)
				return
			}
		}
	}
	if user == nil {
		s.SendError(fmt.Sprintf("Пользователь [%s] не найден", msg.CommandArguments()), msg)
		return
	}
	if user.ID == msg.From.ID {
		s.SendError("Сам себя? O_o", msg)
		return
	}

	currentLevel, err := db.AddWarnLevel(user)
	if err != nil {
		log.Printf("Error in AddWarnLevel: %s", err)
		return
	}

	if currentLevel >= 5 {
		ok, err := s.kickUser(user.ID, msg.Chat, true)
		if err != nil {
			log.Printf("Error in kickUser: %s", err)
			return
		}
		if ok {
			s.SendMessage("Пользователь %s забанен!", msg.Chat.ID, -1)
		}
	}
}

func (s *Server) WarnClear(msg *tgbotapi.Message) {
	isAdmin, err := s.UserIsAdmin(msg.From.ID, msg.Chat)
	if err != nil {
		return
	}
	if !isAdmin {
		s.SendError("Не удалось установить Вашу причастность к администраторам группы!", msg)
		return
	}

	user, err := db.GetUser(msg.CommandArguments())
	if err != nil {
		errStrings := strings.Split(err.Error(), "\n")
		if len(errStrings) > 1 {
			switch errStrings[0] {
			case "User not found":
				s.SendError(fmt.Sprintf("Пользователь %s не найден", msg.CommandArguments()), msg)
				return
			case "Many users":
				s.SendError(fmt.Sprintf("Найдено более одного пользователя, уточните:\n%s", strings.Join(errStrings[1:], "\n")), msg)
				return
			default:
				s.SendError(fmt.Sprintf("Произошла неизвестная ошибка при поиске пользователя: %s", err.Error()), msg)
				return
			}
		}
	}

	if user == nil {
		s.SendError(fmt.Sprintf("Пользователь %s не найден.", msg.CommandArguments()), msg)
		return
	}

	err = db.ClearWarnLevel(user)
	if err != nil {
		log.Printf("Error in WarnClear -> ClearWarnLevel: %s", err)
		return
	}
	s.SendError("Выполнено успешно.", msg)
}

// GetWarnLevel send message with current warning level for user
func (s *Server) GetWarnLevel(msg *tgbotapi.Message) {
	currentLevel, err := db.GetWarnLevel(msg.From)
	if err != nil {
		if err.Error() == "Key not found." {
			s.SendError("Чист душой!", msg)
			return
		}
		log.Printf("Error in GetWarnLevel -> GetWarnLevel: %s", err)
		return
	}
	s.SendError(fmt.Sprintf("Уровень настороженности: %d", currentLevel), msg)
}
