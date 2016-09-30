package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/telegram-bot-api.v4"

	// SQLite 3
	_ "github.com/mattn/go-sqlite3"
)

const (
	databaseFile = "bot.sqlite"
)

// GoSaveMessage is a shell method for goroutine SaveMessage
func GoSaveMessage(msg *tgbotapi.Message) {
	err := SaveMessage(msg)
	if err != nil {
		log.Printf("Error per save message: %s", err.Error())
	}
}

// SaveMessage method save message to database
func SaveMessage(msg *tgbotapi.Message) (err error) {
	if msg == nil {
		return
	}

	var db *sql.DB

	for true {
		db, err = openDatabase()
		if err != nil {
			if err.Error() == "database is locked" {
				time.Sleep(time.Second * 1)
				continue
			}
			return
		}
		break

	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM messages WHERE message_id=?", msg.MessageID)
	if err != nil {
		return
	}

	msgPresent := rows.Next()
	rows.Close()
	var query string

	err = SaveUser(msg.From)
	if err != nil {
		return
	}
	err = SaveChat(msg.Chat)
	if err != nil {
		return
	}
	err = SaveUser(msg.ForwardFrom)
	if err != nil {
		return
	}
	err = SaveChat(msg.ForwardFromChat)
	if err != nil {
		return
	}
	err = SaveMessage(msg.ReplyToMessage)
	if err != nil {
		return
	}

	var forwardFromID int
	var forwardFromChatID int64
	var replyToMessageID int

	if msg.ForwardFrom != nil {
		forwardFromID = msg.ForwardFrom.ID
	}
	if msg.ForwardFromChat != nil {
		forwardFromChatID = msg.ForwardFromChat.ID
	}
	if msg.ReplyToMessage != nil {
		replyToMessageID = msg.ReplyToMessage.MessageID
	}

	messageText := msg.Text
	if strings.Contains(messageText, "\"") {
		strings.Replace(messageText, "\"", "\\\"", -1)
	}

	if !msgPresent { // INSERT
		query = fmt.Sprintf(`INSERT INTO messages
            (message_id, message_from, date, chat, forward_from, forward_from_chat,
                forward_date, reply_to_message, edit_date, text, caption)
            VALUES (%d, %d, %d, %d, %d, %d, %d, %d, %d, "%s", "%s")`,
			msg.MessageID, msg.From.ID, msg.Date, msg.Chat.ID,
			forwardFromID, forwardFromChatID,
			msg.ForwardDate, replyToMessageID,
			msg.EditDate, messageText, msg.Caption)
	} else { // UPDATE
		query = fmt.Sprintf(`UPDATE messages SET
                                message_from=%d, date=%d, chat=%d, forward_from=%d,
                                forward_from_chat=%d, forward_date=%d,
                                reply_to_message=%d, edit_date=%d,
                                text="%s", caption="%s" WHERE message_id=%d`,
			msg.From.ID, msg.Date, msg.Chat.ID,
			forwardFromID, forwardFromChatID,
			msg.ForwardDate, replyToMessageID,
			msg.EditDate, messageText, msg.Caption, msg.MessageID)
	}

	err = runQuery(query)

	return
}

// SaveUser method save user to database
func SaveUser(user *tgbotapi.User) (err error) {
	if user == nil {
		return
	}
	db, err := openDatabase()
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE id=?", user.ID)
	if err != nil {
		return
	}

	userPresent := rows.Next()
	rows.Close()
	var query string
	if !userPresent { // INSERT
		query = fmt.Sprintf(`INSERT INTO users (id, first_name, last_name, username)
                                        VALUES(%d, "%s", "%s", "%s")`,
			user.ID, user.FirstName, user.LastName, user.UserName)
	} else { // UPDATE
		query = fmt.Sprintf(`UPDATE users SET first_name="%s", last_name="%s",
                                username="%s" WHERE id=%d`,
			user.FirstName, user.LastName, user.UserName, user.ID)
	}

	err = runQuery(query)
	return
}

// SaveChat method for save chat to database
func SaveChat(chat *tgbotapi.Chat) (err error) {
	if chat == nil {
		return
	}
	db, err := openDatabase()
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM chats WHERE id=?", chat.ID)
	if err != nil {
		return
	}

	chatPresent := rows.Next()
	rows.Close()

	var query string
	if !chatPresent { // INSERT
		query = fmt.Sprintf(`INSERT INTO chats (id, type, title, first_name,
                            last_name, username) VALUES(%d, "%s", "%s","%s", "%s", "%s")`,
			chat.ID, chat.Type, chat.Title, chat.FirstName, chat.LastName, chat.UserName)
	} else { // UPDATE
		query = fmt.Sprintf(`UPDATE chats SET type="%s", title="%s", first_name="%s", last_name="%s",
                                username="%s" WHERE id=%d`,
			chat.Type, chat.Title, chat.FirstName, chat.LastName, chat.UserName, chat.ID)
	}

	err = runQuery(query)
	return
}

// GetChats returns chat list
func GetChats() (chats []*tgbotapi.Chat, err error) {
	db, err := openDatabase()
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, type, title, first_name, last_name, username FROM chats")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		chat := new(tgbotapi.Chat)
		err = rows.Scan(&chat.ID, &chat.Type, &chat.Title, &chat.FirstName, &chat.LastName, &chat.UserName)
		if err != nil {
			return
		}
		chats = append(chats, chat)
	}

	return
}

// GetMessages returns chat list
func GetMessages() (chats []*tgbotapi.Message, err error) {
	db, err := openDatabase()
	if err != nil {
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT message_id, message_from, date, chat, forward_from, forward_from_chat, forward_date, reply_to_message, edit_date, text, caption FROM chats")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			messageFrom int64
			chat int64
			messageForwardFrom
		)
		msg := new(tgbotapi.Message)
		err = rows.Scan(&msg.MessageID, &)
		if err != nil {
			return
		}
		chats = append(chats, chat)
	}

	return
}

func openDatabase() (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Printf("Error in openning database file %s: %s", databaseFile, err)
		return nil, err
	}

	return
}

func init() {
	db, err := openDatabase()
	if err != nil {
		log.Fatalf("Cannot initialize database!")
	}
	defer db.Close()

	_, err = db.Query("SELECT * FROM messages")
	if err != nil {
		err = runQuery(`CREATE TABLE messages (
                        message_id INTEGER PRIMARY KEY,
                        message_from INTEGER,
                        date INTEGER,
                        chat INTEGER,
                        forward_from INTEGER,
                        forward_from_chat INTEGER,
                        forward_date INTEGER,
                        reply_to_message INTEGER,
                        edit_date INTEGER,
                        text STRING,
                        caption STRING)`)
		if err != nil {
			log.Fatalf("Error in create table messages: %s!", err)
		}
	}
	_, err = db.Query("SELECT * FROM users")
	if err != nil {
		err = runQuery(`CREATE TABLE users (
                        id INTEGER PRIMARY KEY,
                        first_name STRING,
                        last_name STRING,
                        username STRING)`)
		if err != nil {
			log.Fatalf("Error in create table users: %s!", err)
		}
	}
	_, err = db.Query("SELECT * FROM chats")
	if err != nil {
		err = runQuery(`CREATE TABLE chats (
                        id INTEGER PRIMARY KEY,
                        type STRING,
                        title STRING,
                        first_name STRING,
                        last_name STRING,
                        username STRING)`)
		if err != nil {
			log.Fatalf("Error in create table chats: %s!", err)
		}
	}
}

func runQuery(query string) (err error) {
	db, err := openDatabase()
	if err != nil {
		return
	}
	defer db.Close()

	stm, err := db.Prepare(query)
	if err != nil {
		log.Printf("Error in prepare query [%s]", query)
		return err
	}
	defer stm.Close()
	_, err = stm.Exec()
	if err != nil {
		log.Printf("Error in exec query [%s]", query)
		return err
	}

	return
}
