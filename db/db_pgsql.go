package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/telegram-bot-api.v4"

	// import for PostgreSQL
	_ "github.com/lib/pq"
)

var (
	pgsql *sql.DB
)

// InitPGSQL function initial postgresql connection
func InitPGSQL(connString string) {
	var err error
	pgsql, err = sql.Open("postgres", connString)
	if err != nil {
		log.Fatalf("Error in connection to database: %s", err)
		return
	}
	pgsql.SetMaxIdleConns(-1)
	pgsql.SetMaxOpenConns(5)

}

func execQuery(queryStr string) (err error) {
	if pgsql == nil {
		return fmt.Errorf("database connection pointer is nil")
	}
	_, err = pgsql.Exec(queryStr)
	if err != nil {
		log.Printf("Error in query \"%s\"", queryStr)
	}
	return
}

func selectQuery(queryStr string) (rows *sql.Rows, err error) {
	if pgsql == nil {
		return nil, fmt.Errorf("database connection pointer is nil")
	}
	rows, err = pgsql.Query(queryStr)
	if err != nil {
		log.Printf("Error in query \"%s\"", queryStr)
	}
	return
}

func selectOneQuery(queryStr string) (row *sql.Row, err error) {
	if pgsql == nil {
		return nil, fmt.Errorf("database connection pointer is nil")
	}
	row = pgsql.QueryRow(queryStr)
	return
}

func recordIsExist(table string, id int) (result bool) {
	queryStr := fmt.Sprintf("SELECT id FROM %s WHERE id=%d", table, id)
	row, err := selectOneQuery(queryStr)
	if err != nil {
		return
	}
	var tempid int
	err = row.Scan(&tempid)
	if err == sql.ErrNoRows {
		return false
	} else if err != nil {
		log.Printf("Error in scan results from PGSQL: %s", err)
		return false
	}
	return true
}

// GetUserPGSQL get user by username or first and last name
func GetUserPGSQL(username string) (user *tgbotapi.User, err error) {
	var queryStr string
	if username[0] == '@' {
		queryStr = fmt.Sprintf("SELECT * FROM users WHERE user_name='%s'", username[1:])
	} else {
		argList := strings.Split(username, " ")
		switch len(argList) {
		case 1:
			queryStr = fmt.Sprintf("SELECT * FROM users WHERE first_name='%s'", argList[0])
		case 2:
			queryStr = fmt.Sprintf("SELECT * FROM users WHERE first_name='%s' AND last_name='%s'", argList[0], argList[1])
		default:
			return nil, fmt.Errorf("User not found\n%s", username)
		}
	}
	row, err := selectOneQuery(queryStr)
	if err != nil {
		return nil, err
	}

	user = new(tgbotapi.User)
	err = row.Scan(user.ID, user.FirstName, user.LastName, user.UserName)
	if err != nil {
		return nil, err
	}

	return
}

func getUserByID(id int) (user *tgbotapi.User, err error) {
	queryStr := fmt.Sprintf("SELECT id, first_name, last_name, user_name FROM users WHERE id=%d", id)
	row, err := selectOneQuery(queryStr)
	if err != nil {
		return nil, err
	}

	user = new(tgbotapi.User)
	err = row.Scan(user.ID, user.FirstName, user.LastName, user.UserName)
	if err != nil {
		return nil, err
	}

	return
}

// SaveUserPGSQL method save user to database
func SaveUserPGSQL(user *tgbotapi.User) (err error) {
	var tempUser *tgbotapi.User
	tempUser, err = getUserByID(user.ID)
	if err != nil && err != sql.ErrNoRows {
		return
	} else if err == sql.ErrNoRows {
		tempUser = new(tgbotapi.User)
	}

	if tempUser.String() == user.String() {
		return
	}

	var queryStr string
	if tempUser.ID != 0 {
		queryStr = fmt.Sprintf("UPDATE users SET first_name='%s', last_name='%s', user_name='%s' WHERE id=%d",
			user.FirstName, user.LastName, user.UserName, user.ID)
	} else {
		queryStr = fmt.Sprintf("INSERT INTO users (id, first_name, last_name, user_name) VALUES (%d, '%s', '%s', '%s')",
			user.ID, user.FirstName, user.LastName, user.UserName)
	}

	err = execQuery(queryStr)

	return
}

// GetFilePGSQL returns file json from couchbase
func GetFilePGSQL(fileID string, chatID int64) (f *tgbotapi.File, err error) {
	queryStr := fmt.Sprintf("SELECT id,file_path,file_size FROM files WHERE id='%s' AND chat_id=%d", fileID, chatID)
	row, err := selectOneQuery(queryStr)
	if err != nil {
		return
	}
	f = new(tgbotapi.File)
	err = row.Scan(f.FileID, f.FilePath, f.FileSize)
	if err != nil {
		return nil, err
	}
	return
}

// SaveFilePGSQL method save user to database
func SaveFilePGSQL(file *tgbotapi.File, chatID int64) (err error) {
	var tempFile *tgbotapi.File
	tempFile, err = GetFilePGSQL(file.FileID, chatID)
	if err == sql.ErrNoRows {
		tempFile = new(tgbotapi.File)
	} else if err != nil {
		return
	}

	if tempFile.Link("test") == file.Link("test") {
		return
	}

	var queryStr string
	if tempFile.FileID == "" {
		queryStr = fmt.Sprintf("INSERT INTO files (id, chat_id, file_path, file_size) VALUES ('%s', %d, '%s', %d)",
			file.FileID, chatID, file.FilePath, file.FileSize)
	} else {
		queryStr = fmt.Sprintf("UPDATE files SET chat_id=%d, file_path='%s', file_size=%d WHERE id='%s'",
			chatID, file.FilePath, file.FileSize, file.FileID)
	}
	err = execQuery(queryStr)

	return
}

// GetCensLevelPGSQL function returns censore level for user
func GetCensLevelPGSQL(user *tgbotapi.User) (currentLevel int, err error) {
	queryStr := fmt.Sprintf("SELECT user_id, level, year FROM censlevels WHERE user_id=%d AND year=%d", user.ID, time.Now().Year())
	row, err := selectOneQuery(queryStr)
	if err != nil {
		return
	}

	cens := CensLevel{}
	err = row.Scan(&cens.ID, &cens.Level, &cens.Year)
	if err == sql.ErrNoRows {
		return
	} else if err != nil {
		return
	}

	return cens.Level, nil
}

// SetCensLevelPGSQL function sets level for user
func SetCensLevelPGSQL(user *tgbotapi.User, setlevel int) (err error) {
	var queryStr string
	_, err = GetCensLevelPGSQL(user)
	year := time.Now().Year()
	if err == sql.ErrNoRows {
		queryStr = fmt.Sprintf("INSERT INT censlevels (user_id, level, year) VALUES(%d, %d, %d)", user.ID, setlevel, year)
	} else if err != nil {
		return
	} else {
		queryStr = fmt.Sprintf("UPDATE censlevels SET level=%d WHERE user_id=%d NAD year=%d", setlevel, user.ID, year)
	}
	err = execQuery(queryStr)

	return
}

// ClearCensLevelPGSQL remove document from bucket
func ClearCensLevelPGSQL(user *tgbotapi.User) (err error) {
	queryStr := fmt.Sprintf("DELETE FROM censlevels WHERE user_id=%d AND year=%d", user.ID, time.Now().Year())
	err = execQuery(queryStr)
	return
}

// AddCensLevelPGSQL added +1 to cens level in year
func AddCensLevelPGSQL(user *tgbotapi.User) (currentLevel int, err error) {
	var queryStr string
	currentLevel, err = GetCensLevelPGSQL(user)
	year := time.Now().Year()
	if err == sql.ErrNoRows {
		queryStr = fmt.Sprintf("INSERT INT censlevels (user_id, level, year) VALUES(%d, %d, %d)", user.ID, 1, year)
	} else if err != nil {
		return
	} else {
		currentLevel++
		queryStr = fmt.Sprintf("UPDATE censlevels SET level=%d WHERE user_id=%d NAD year=%d", currentLevel, user.ID, year)
	}
	err = execQuery(queryStr)

	return
}

// SaveChatPGSQL method for save chat to database
func SaveChatPGSQL(chat *tgbotapi.Chat, forward bool) (err error) {
	var tempChat *tgbotapi.Chat

	tempChat, err = getChat(chat.ID)
	if err == sql.ErrNoRows {
		tempChat = new(tgbotapi.Chat)
	} else if err != nil {
		return
	}

	if *tempChat == *chat {
		return
	}

	var queryStr string
	ftype := typeChat
	if forward {
		ftype = typeForwardType
	}
	if tempChat.ID == 0 {
		queryStr = fmt.Sprintf(`INSERT INTO chats (id, first_name, last_name, title, user_name, chat_type, ftype)
                                VALUES(%d, '%s', '%s', '%s', '%s', '%s', '%s')`,
			chat.ID, chat.FirstName, chat.LastName, chat.Title, chat.UserName, chat.Type, ftype)
	} else {
		queryStr = fmt.Sprintf("UPDATE chats SET first_name='%s', last_name='%s', title='%s', user_name='%s', chat_type='%s' WHERE id=%d",
			chat.FirstName, chat.LastName, chat.Title, chat.UserName, chat.Type, chat.ID)
	}
	err = execQuery(queryStr)

	return
}

func getChat(id int64) (chat *tgbotapi.Chat, err error) {
	queryStr := fmt.Sprintf("SELECT id, first_name, last_name, title, user_name, chat_type FROM chats WHERE id=%d", id)
	row, err := selectOneQuery(queryStr)
	if err != nil {
		return
	}

	chat = new(tgbotapi.Chat)
	err = row.Scan(chat.ID, chat.FirstName, chat.LastName, chat.Title, chat.UserName, chat.Type)
	if err != nil {
		return nil, err
	}

	return
}

// GetChatsPGSQL returns chat list
func GetChatsPGSQL() (chats []*tgbotapi.Chat, err error) {
	queryStr := fmt.Sprintf("SELECT id, first_name, last_name, title, user_name, chat_type FROM chats WHERE ftype='%s'", typeChat)
	rows, err := selectQuery(queryStr)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		chat := new(tgbotapi.Chat)
		err = rows.Scan(chat.ID, chat.FirstName, chat.LastName, chat.Title, chat.UserName, chat.Type)
		if err != nil {
			return []*tgbotapi.Chat{}, err
		}
		chats = append(chats, chat)
	}

	return
}

func saveEntity(msgID int, ent tgbotapi.MessageEntity) (err error) {
	userID := 0
	if ent.User != nil {
		err = SaveUserPGSQL(ent.User)
		if err != nil {
			return
		}
		userID = ent.User.ID
	}

	queryStr := fmt.Sprintf(`INSERT INTO entities (msg_id, type, offset, length, url, user_id)
                            VALUES (%d, '%s', %d, %d, '%s', %d)`,
		msgID, ent.Type, ent.Offset, ent.Length, ent.URL, userID)
	err = execQuery(queryStr)
	return
}

func saveEntities(msgID int, ents *[]tgbotapi.MessageEntity) (err error) {
	if ents == nil {
		return
	}

	queryStr := fmt.Sprintf("DELETE FROM entities WHERE msg_id=%d", msgID)
	execQuery(queryStr)
	for _, ent := range *ents {
		err = saveEntity(msgID, ent)
	}

	return
}

func saveAudio(audio *tgbotapi.Audio) (err error) {
	if audio == nil {
		return
	}

	queryStr := fmt.Sprintf("DELETE FROM audio WHERE file_id='%s'", audio.FileID)
	execQuery(queryStr)

	queryStr = fmt.Sprintf(`INSERT INTO audio (file_id, duration, performer, title, mime, size)
                            VALUES('%s', %d, '%s', '%s', '%s', %d)`,
		audio.FileID, audio.Duration, audio.Performer,
		audio.Title, audio.MimeType, audio.FileSize)
	err = execQuery(queryStr)
	return
}

func savePhotoSize(msgID int, photo *tgbotapi.PhotoSize, newPhoto bool) (err error) {
	if photo == nil {
		return
	}

	queryStr := fmt.Sprintf("DELETE FROM photosize WHERE file_id='%s'", photo.FileID)
	execQuery(queryStr)

	if msgID == 0 {
		queryStr = fmt.Sprintf(`INSERT INTO photosize (file_id, width, height, size)
        VALUES('%s', %d, %d, %d)`,
			photo.FileID, photo.Width, photo.Height, photo.FileSize)
	} else {
		queryStr = fmt.Sprintf(`INSERT INTO photosize (file_id, width, height, size, msg_id, new)
        VALUES('%s', %d, %d, %d, '%d', %t)`,
			photo.FileID, photo.Width, photo.Height, photo.FileSize, msgID, newPhoto)

	}
	err = execQuery(queryStr)
	return
}

func savePhoto(msgID int, photos *[]tgbotapi.PhotoSize, newPhoto bool) (err error) {
	if photos == nil {
		return
	}

	for _, photo := range *photos {
		err = savePhotoSize(msgID, &photo, newPhoto)
		if err != nil {
			return
		}
	}

	return
}

func saveDocument(document *tgbotapi.Document) (err error) {
	if document == nil {
		return
	}

	queryStr := fmt.Sprintf("DELETE FROM documents WHERE file_id='%s'", document.FileID)
	execQuery(queryStr)

	thumbnailID := ""
	if document.Thumbnail != nil {
		err = savePhotoSize(0, document.Thumbnail, false)
		if err != nil {
			return
		}
		thumbnailID = document.Thumbnail.FileID
	}

	queryStr = fmt.Sprintf(`INSERT INTO documents (file_id, thumbnail_id, file_name, mime, size)
                            VALUES('%s', '%s', '%s', '%s', %d)`,
		document.FileID, thumbnailID,
		document.FileName, document.MimeType, document.FileSize)
	err = execQuery(queryStr)
	return
}

func saveSticker(document *tgbotapi.Sticker) (err error) {
	if document == nil {
		return
	}

	queryStr := fmt.Sprintf("DELETE FROM stickers WHERE file_id='%s'", document.FileID)
	execQuery(queryStr)

	thumbnailID := ""
	if document.Thumbnail != nil {
		err = savePhotoSize(0, document.Thumbnail, false)
		if err != nil {
			return
		}
		thumbnailID = document.Thumbnail.FileID
	}

	queryStr = fmt.Sprintf(`INSERT INTO stickers (file_id, thumbnail_id, width, height, emoji, size)
                            VALUES('%s', '%s', %d, %d, '%s', %d)`,
		document.FileID, thumbnailID, document.Width, document.Height,
		document.Emoji, document.FileSize)
	err = execQuery(queryStr)
	return
}

func saveVideo(document *tgbotapi.Video) (err error) {
	if document == nil {
		return
	}

	queryStr := fmt.Sprintf("DELETE FROM video WHERE file_id='%s'", document.FileID)
	execQuery(queryStr)

	thumbnailID := ""
	if document.Thumbnail != nil {
		err = savePhotoSize(0, document.Thumbnail, false)
		if err != nil {
			return
		}
		thumbnailID = document.Thumbnail.FileID
	}

	queryStr = fmt.Sprintf(`INSERT INTO video (file_id, thumbnail_id, width, height, duration, mime, size)
                            VALUES('%s', '%s', %d, %d, %d, '%s', %d)`,
		document.FileID, thumbnailID, document.Width, document.Height,
		document.Duration, document.MimeType, document.FileSize)
	err = execQuery(queryStr)
	return
}

func saveVoice(document *tgbotapi.Voice) (err error) {
	if document == nil {
		return
	}

	queryStr := fmt.Sprintf("DELETE FROM voices WHERE file_id='%s'", document.FileID)
	execQuery(queryStr)

	queryStr = fmt.Sprintf(`INSERT INTO voices (file_id, duration, mime, size)
                            VALUES('%s', %d, '%s', %d)`,
		document.FileID, document.Duration, document.MimeType, document.FileSize)
	err = execQuery(queryStr)
	return
}

func saveMessageDeps(msg *tgbotapi.Message) (err error) {
	if msg.From != nil {
		err = SaveUserPGSQL(msg.From)
		if err != nil {
			return
		}
	}

	if msg.ForwardFrom != nil {
		err = SaveUserPGSQL(msg.ForwardFrom)
		if err != nil {
			return
		}
	}

	if msg.Chat != nil {
		err = SaveChatPGSQL(msg.Chat, false)
		if err != nil {
			return
		}
	}
	if msg.ForwardFromChat != nil {
		err = SaveChatPGSQL(msg.ForwardFromChat, true)
		if err != nil {
			return
		}
	}
	if msg.ReplyToMessage != nil {
		err = SaveMessagePGSQL(msg.ReplyToMessage)
		if err != nil {
			return
		}
	}

	if msg.Entities != nil {
		err = saveEntities(msg.MessageID, msg.Entities)
		if err != nil {
			return
		}
	}

	if msg.Audio != nil {
		err = saveAudio(msg.Audio)
		if err != nil {
			return
		}
	}

	if msg.Document != nil {
		err = saveDocument(msg.Document)
		if err != nil {
			return
		}
	}

	if msg.Photo != nil {
		err = savePhoto(msg.MessageID, msg.Photo, false)
		if err != nil {
			return
		}
	}

	if msg.Sticker != nil {
		err = saveSticker(msg.Sticker)
		if err != nil {
			return
		}
	}

	if msg.Video != nil {
		err = saveVideo(msg.Video)
		if err != nil {
			return
		}
	}
	if msg.Voice != nil {
		err = saveVoice(msg.Voice)
		if err != nil {
			return
		}
	}

	if msg.NewChatMember != nil {
		err = SaveUserPGSQL(msg.NewChatMember)
		if err != nil {
			return
		}
	}

	if msg.LeftChatMember != nil {
		err = SaveUserPGSQL(msg.LeftChatMember)
		if err != nil {
			return
		}
	}

	if msg.NewChatPhoto != nil {
		err = savePhoto(msg.MessageID, msg.NewChatPhoto, true)
		if err != nil {
			return
		}
	}

	if msg.PinnedMessage != nil {
		err = SaveMessage(msg.PinnedMessage)
		if err != nil {
			return
		}
	}

	return
}

// SaveMessagePGSQL function save message to database
func SaveMessagePGSQL(msg *tgbotapi.Message) (err error) {
	err = saveMessageDeps(msg)
	if err != nil {
		return
	}

	queryStr := fmt.Sprintf("DELETE FROM messages WHERE id=%d", msg.MessageID)
	execQuery(queryStr)

	var (
		fromID             int
		chatID             int64
		fFromID            int
		fChatID            int64
		replyMsgID         int
		audioID            string
		doucementID        string
		stickerID          string
		videoID            string
		voiceID            string
		contactPhoneNumber string
		contactFirstName   string
		contactLastName    string
		contactUserID      int
		locationLongitude  float64
		locationLatitude   float64
		venueLongitude     float64
		venueLatitude      float64
		venueTitle         string
		venueAddress       string
		venueFoursqueareID string
		newChatID          int
		leftChatID         int
		pinnedMessageID    int
	)

	if msg.From != nil {
		fromID = msg.From.ID
	}
	if msg.Chat != nil {
		chatID = msg.Chat.ID
	}
	if msg.ForwardFrom != nil {
		fFromID = msg.ForwardFrom.ID
	}
	if msg.ForwardFromChat != nil {
		fChatID = msg.ForwardFromChat.ID
	}
	if msg.ReplyToMessage != nil {
		replyMsgID = msg.ReplyToMessage.MessageID
	}
	if msg.Audio != nil {
		audioID = msg.Audio.FileID
	}
	if msg.Document != nil {
		doucementID = msg.Document.FileID
	}
	queryStr = fmt.Sprintf(`INSERT INTO messages (  id,
                                                    from_id,
                                                    date,
                                                    chat_id,
                                                    forward_from_id,
                                                    forward_from_chat_id,
                                                    forward_date,
                                                    reply_to_message_id,
                                                    edit_date,
                                                    text,
                                                    audio_id,
                                                    document_id,
                                                    sticker_id,
                                                    video_id,
                                                    voice_id,
                                                    caption,
                                                    contact_phone_number,
                                                    contact_first_name
                                                    contact_last_name,
                                                    contact_user_id,
                                                    location_longitude,
                                                    location_latitude,
                                                    venue_longitude,
                                                    venue_latitude,
                                                    venue_title,
                                                    venue_address,
                                                    venue_foursquare_id,
                                                    new_chat_member_id,
                                                    left_chat_member_id,
                                                    new_chat_title,
                                                    delete_chat_photo,
                                                    group_chat_created,
                                                    super_group_chat_created,
                                                    channel_chat_created,
                                                    migrate_to_chat_id,
                                                    migrate_from_chat_id,
                                                    pinned_message_id)
                            VALUES( %d,
                                    %d,
                                    %d,
                                    %d,
                                    %d,
                                    %d,
                                    %d,
                                    %d,
                                    %d,
                                    '%s',
                                    '%s',
                                    '%s',
                                    '%s',
                                    '%s',
                                    '%s',
                                    '%s',
                                    '%s',
                                    '%s',
                                    '%s',
                                    %d,
                                    %f,%f,
                                    %f, %f, '%s', '%s', '%s',
                                    %d,
                                    %d,
                                    '%s',
                                    %t,
                                    %t,
                                    %t,
                                    %t,
                                    %d,
                                    %d,
                                    %d)`,
		msg.MessageID,
		fromID,
		msg.Date,
		chatID,
		fFromID,
		fChatID,
		msg.ForwardDate,
		replyMsgID,
		msg.EditDate,
		msg.Text,
		audioID,
		doucementID,
		stickerID,
		videoID,
		voiceID,
		msg.Caption,
		contactPhoneNumber,
		contactFirstName,
		contactLastName,
		contactUserID,
		locationLongitude,
		locationLatitude,
		venueLongitude,
		venueLatitude,
		venueTitle,
		venueAddress,
		venueFoursqueareID,
		newChatID,
		leftChatID,
		msg.NewChatTitle,
		msg.DeleteChatPhoto,
		msg.GroupChatCreated,
		msg.SuperGroupChatCreated,
		msg.ChannelChatCreated,
		msg.MigrateToChatID,
		msg.MigrateFromChatID,
		pinnedMessageID)
	err = execQuery(queryStr)

	return
}

// GetMessagesPGSQL returns chat list
func GetMessagesPGSQL(chatID int64) (messages []*tgbotapi.Message, err error) {
	// queryStr := fmt.Sprintf("SELECT * FROM messages WHERE chat_id=%d", chatID)
	// rows, err := selectQuery(queryStr)
	// if err != nil {
	// 	return
	// }
	// defer rows.Close()
	//
	// for rows.Next() {
	// 	msg := new(tgbotapi.Message)
	// 	var (
	// 		audioID           string
	// 		userID            int
	// 		chatID            int64
	// 		fUserID           int
	// 		fChatID           int64
	// 		replyMsgID        int
	// 		stickerID         string
	// 		videoID           string
	// 		voiceID           string
	// 		contactID         int
	// 		longitude         float64
	// 		latitude          float64
	// 		venueLongitude    float64
	// 		venueLatitude     float64
	// 		venueTitle        string
	// 		venueAddress      string
	// 		venueFoursquareID string
	// 		newChatMemberID   int
	// 		leftChatMember    int
	// 		pinnedMessageID   int
	// 	)
	// 	err = rows.Scan(msg.ID, msg)
	// 	if err != nil {
	// 		return []*tgbotapi.Message{}, err
	// 	}
	// 	messages = append(messages, msg)
	// }
	return
}
