package db

import (
	"encoding/json"
	"fmt"
	"log"

	couchbase "github.com/couchbase/gocb"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	databaseFile     = "bot.sqlite"
	couchbaseCluster = "couchbase://172.16.32.81"
	couchbaseBucket  = "RussianFedoraBot"
	couchbaseSecret  = "3510"
)

var (
	cluster *couchbase.Cluster
	bucket  *couchbase.Bucket
)

func initCouchbase(couchbaseCluster, couchbaseBucket, couchbaseSecret string) {
	cluster, err := couchbase.Connect(couchbaseCluster)
	if err != nil {
		log.Fatalf("Cannot connect to cluster: %s", err)
	}
	bucket, err = cluster.OpenBucket(couchbaseBucket, couchbaseSecret)
	if err != nil {
		log.Fatalf("Cannot open bucket: %s", err)
	}
}

func init() {
	initCouchbase(couchbaseCluster, couchbaseBucket, couchbaseSecret)
}

// GoSaveMessage is a shell method for goroutine SaveMessage
func GoSaveMessage(msg *tgbotapi.Message) {
	err := SaveMessage(msg)
	if err != nil {
		log.Printf("Error per save message: %s", err.Error())
	}
}

// SaveMessage method save message to database
func SaveMessage(msg *tgbotapi.Message) (err error) {
	key := fmt.Sprintf("message:%d:%d", msg.Chat.ID, msg.MessageID)

	type couchmessage struct {
		tgbotapi.Message
		Type string `json:"type"`
	}
	cMsg := couchmessage{}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &cMsg)
	cMsg.Type = "message"

	_, err = bucket.Upsert(key, &cMsg, 0)

	if msg.Chat != nil {
		err = SaveChat(msg.Chat)
	}
	if msg.ForwardFrom != nil {
		err = SaveUser(msg.ForwardFrom)
	}
	if msg.ForwardFromChat != nil {
		err = SaveChat(msg.ForwardFromChat)
	}
	if msg.ReplyToMessage != nil {
		err = SaveMessage(msg.ReplyToMessage)
	}
	if msg.From != nil {
		err = SaveUser(msg.From)
	}
	if msg.NewChatMember != nil {
		err = SaveUser(msg.NewChatMember)
	}

	return
}

// SaveUser method save user to database
func SaveUser(user *tgbotapi.User) (err error) {
	key := fmt.Sprintf("user:%d", user.ID)

	type couchuser struct {
		tgbotapi.User
		Type string `json:"type"`
	}
	cUser := couchuser{}

	data, err := json.Marshal(user)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &cUser)
	cUser.Type = "user"

	_, err = bucket.Upsert(key, &cUser, 0)
	return
}

// SaveChat method for save chat to database
func SaveChat(chat *tgbotapi.Chat) (err error) {
	key := fmt.Sprintf("chat:%d", chat.ID)

	type couchchat struct {
		tgbotapi.Chat
		Type string `json:"type"`
	}
	cChat := couchchat{}

	data, err := json.Marshal(chat)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &cChat)
	cChat.Type = "chat"

	_, err = bucket.Upsert(key, cChat, 0)
	return
}

// GetChats returns chat list
func GetChats() (chats []*tgbotapi.Chat, err error) {
	//query := couchbase.NewN1qlQuery("SELECT first_name, id, last_name, title, type, username FROM RussianFedoraBot WHERE type='chat'")

	type couchchat struct {
		Msg tgbotapi.Chat `json:"RussianFedoraBot"`
	}

	query := couchbase.NewN1qlQuery("SELECT * FROM RussianFedoraBot WHERE type='chat'")
	res, err := bucket.ExecuteN1qlQuery(query, nil)
	if err != nil {
		return
	}

	//var data interface{}

	chat := couchchat{}
	for res.Next(&chat) {

		data, err := json.Marshal(chat.Msg)
		if err != nil {
			log.Printf("Error in marshal GetChats: %s", err)
			continue
		}
		oChat := new(tgbotapi.Chat)
		err = json.Unmarshal(data, oChat)
		if err != nil {
			log.Printf("Error in unmarshal GetChats: %s", err)
			continue
		}
		chats = append(chats, oChat)
	}

	return
}

// GetMessages returns chat list
func GetMessages(chatID int64) (messages []*tgbotapi.Message, err error) {
	type couchmsg struct {
		Msg tgbotapi.Message `json:"RussianFedoraBot"`
	}

	queryStr := fmt.Sprintf("SELECT * FROM RussianFedoraBot WHERE type='message' AND chat.id=%d ORDER BY date", chatID)
	query := couchbase.NewN1qlQuery(queryStr)
	res, err := bucket.ExecuteN1qlQuery(query, nil)
	if err != nil {
		return
	}

	//var data interface{}

	chat := couchmsg{}
	for res.Next(&chat) {
		data, err := json.Marshal(chat.Msg)
		if err != nil {
			log.Printf("Error in marshal GetMessages: %s", err)
			continue
		}
		oMsg := new(tgbotapi.Message)
		err = json.Unmarshal(data, oMsg)
		if err != nil {
			log.Printf("Error in unmarshal GetMessages: %s", err)
			continue
		}
		messages = append(messages, oMsg)
	}

	return
}
