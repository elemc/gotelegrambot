package db

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	couchbase "github.com/couchbase/gocb"
	"gopkg.in/telegram-bot-api.v4"
)

const (
	typeChat        = "chat"
	typeForwardType = "forward-type"
)

var (
	caches       Caches
	useCouchbase bool
	usePGSQL     bool
)

// CensLevel main struct for records censlevel:year:id
type CensLevel struct {
	ID    int `json:"user_id"`
	Level int `json:"level"`
	Year  int `json:"year"`
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
	go AddedDateToCaches(msg.Chat.ID, msg.Time())
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
	if err != nil {
		return
	}
	cMsg.Type = "message"

	_, err = bucket.Upsert(key, &cMsg, 0)

	if msg.Chat != nil {
		err = SaveChat(msg.Chat, false)
	}
	if msg.ForwardFrom != nil {
		err = SaveUser(msg.ForwardFrom)
	}
	if msg.ForwardFromChat != nil {
		err = SaveChat(msg.ForwardFromChat, true)
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
	if useCouchbase {
		err = SaveUserCouchbase(user)
	} else if usePGSQL {
		err = SaveUserPGSQL(user)
	}
	return
}

// SaveFile method save user to database
func SaveFile(file *tgbotapi.File, chatID int64) (err error) {
	if useCouchbase {
		err = SaveFileCouchbase(file, chatID)
	} else if usePGSQL {
		err = SaveFilePGSQL(file, chatID)
	}
	return
}

// GetCensLevel function returns censore level for user
func GetCensLevel(user *tgbotapi.User) (currentLevel int, err error) {
	if useCouchbase {
		return GetCensLevelCouchbase(user)
	} else if usePGSQL {
		return GetCensLevelPGSQL(user)
	}
	return
}

// SetCensLevel function sets level for user
func SetCensLevel(user *tgbotapi.User, setlevel int) (err error) {
	if useCouchbase {
		err = SetCensLevelCouchbase(user, setlevel)
	} else if usePGSQL {
		err = SetCensLevelPGSQL(user, setlevel)
	}
	return
}

// ClearCensLevel remove document from bucket
func ClearCensLevel(user *tgbotapi.User) (err error) {
	if useCouchbase {
		err = ClearCensLevelCouchbase(user)
	} else if usePGSQL {
		err = ClearCensLevelPGSQL(user)
	}
	return
}

// AddCensLevel added +1 to cens level in year
func AddCensLevel(user *tgbotapi.User) (currentLevel int, err error) {
	currentLevel, err = GetCensLevel(user)
	if err != nil {
		currentLevel = 1
		err = SetCensLevel(user, currentLevel)
		return
	}
	currentLevel++
	err = SetCensLevel(user, currentLevel)

	return
}

// GetFile returns file json from couchbase
func GetFile(fileID string, chatID int64) (f *tgbotapi.File, err error) {
	if useCouchbase {
		f, err = GetFileCouchbase(fileID, chatID)
	} else if usePGSQL {
		f, err = GetFilePGSQL(fileID, chatID)
	}
	return
}

// SaveChat method for save chat to database
func SaveChat(chat *tgbotapi.Chat, forward bool) (err error) {
	if useCouchbase {
		err = SaveChatCouchbase(chat, forward)
	} else if usePGSQL {
		err = SaveChatPGSQL(chat, forward)
	}
	return
}

// GetChats returns chat list
func GetChats() (chats []*tgbotapi.Chat, err error) {
	if useCouchbase {
		chats, err = GetChatsCouchbase()
	} else if usePGSQL {
		chats, err = GetChatsPGSQL()
	}
	return
}

// GetMessages returns chat list
func GetMessages(chatID int64) (messages []*tgbotapi.Message, err error) {
	type couchmsg struct {
		Msg tgbotapi.Message `json:"bot"`
	}

	queryStr := fmt.Sprintf("SELECT * FROM %s AS bot WHERE type='message' AND chat.id=%d ORDER BY date", bucketName, chatID)
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

// GetMessagesByDate returns chat list on date
func GetMessagesByDate(chatID int64, beginTime, endTime time.Time) (messages []*tgbotapi.Message, err error) {
	type couchmsg struct {
		Msg tgbotapi.Message `json:"bot"`
	}

	queryStr := fmt.Sprintf("SELECT * FROM %s AS bot WHERE type='message' AND chat.id=%d AND date >= %d AND date <= %d ORDER BY date", bucketName, chatID, beginTime.Unix(), endTime.Unix())
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

// GetUsers returns chat list
func GetUsers() (users []*tgbotapi.User, err error) {
	type couchuser struct {
		User tgbotapi.User `json:"bot"`
	}

	query := couchbase.NewN1qlQuery(fmt.Sprintf("SELECT * FROM %s AS bot WHERE type='user'", bucketName))
	res, err := bucket.ExecuteN1qlQuery(query, nil)
	if err != nil {
		return
	}

	//var data interface{}

	user := couchuser{}
	for res.Next(&user) {

		data, err := json.Marshal(user.User)
		if err != nil {
			log.Printf("Error in marshal GetUsers: %s", err)
			continue
		}
		oUser := new(tgbotapi.User)
		err = json.Unmarshal(data, oUser)
		if err != nil {
			log.Printf("Error in unmarshal GetUsers: %s", err)
			continue
		}
		users = append(users, oUser)
	}

	return
}

func getDates(chatID int64, beginDate, endDate int64) (result []time.Time, err error) {
	type couchdate struct {
		Date int64 `json:"date"`
	}

	var dateWhere string
	if beginDate != 0 || endDate != 0 {
		dateWhere = fmt.Sprintf(" AND date >= %d AND date <= %d", beginDate, endDate)
	}

	queryStr := fmt.Sprintf("SELECT date FROM %s WHERE type='message' AND chat.id=%d %s ORDER BY date", bucketName, chatID, dateWhere)
	query := couchbase.NewN1qlQuery(queryStr)
	res, err := bucket.ExecuteN1qlQuery(query, nil)
	if err != nil {
		return
	}

	date := couchdate{}
	for res.Next(&date) {
		tDate := time.Unix(date.Date, 0)
		result = append(result, tDate)
	}
	return
}

func appendIfNotFound(list []string, s string) []string {
	found := false
	for _, value := range list {
		if value == s {
			found = true
			break
		}
	}

	if !found {
		list = append(list, s)
	}
	return list
}

func appendIfNotFoundMonth(list []time.Month, s time.Month) []time.Month {
	found := false
	for _, value := range list {
		if value == s {
			found = true
			break
		}
	}

	if !found {
		list = append(list, s)
	}
	return list
}

func appendIfNotFoundInt(list []int, s int) []int {
	found := false
	for _, value := range list {
		if value == s {
			found = true
			break
		}
	}

	if !found {
		list = append(list, s)
	}
	return list
}

// GetYears function returns years msg date from chat messages
func GetYears(chatID int64) (result []string, err error) {
	years := getCache(chatID).Years
	if len(years) != 0 {
		sort.Strings(years)
		return years, nil
	}
	listDates, err := getDates(chatID, 0, 0)
	if err != nil {
		return
	}
	for _, t := range listDates {
		go AddedDateToCaches(chatID, t)
		s := strconv.Itoa(t.Year())
		result = appendIfNotFound(result, s)
	}
	return
}

// GetMonthList function returns month list msg date from chat messages and year
func GetMonthList(chatID int64, year int) (result []time.Month, err error) {
	cache := getCache(chatID)
	if list, ok := cache.MonthsByYear[year]; ok {
		if len(list) > 0 {
			return sortMonths(list), nil
		}
	}
	beginDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local).Unix()
	endDate := time.Date(year, 12, 31, 23, 59, 59, 100, time.Local).Unix()
	listDates, err := getDates(chatID, beginDate, endDate)
	if err != nil {
		return
	}
	for _, t := range listDates {
		if t.Year() != year {
			continue
		}

		result = appendIfNotFoundMonth(result, t.Month())
	}
	return

}

// GetDates function returns month list msg date from chat messages and year
func GetDates(chatID int64, year int, month int) (result []int, err error) {
	result = getDays(chatID, year, time.Month(month))
	if len(result) > 0 {
		return
	}
	beginTime := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	beginDate := beginTime.Unix()
	endDate := time.Date(year, time.Month(month), 32, 23, 59, 59, 100, time.Local).Unix()
	listDates, err := getDates(chatID, beginDate, endDate)
	if err != nil {
		return
	}
	for _, t := range listDates {
		if t.Year() == year && t.Month() == time.Month(month) {
			result = appendIfNotFoundInt(result, t.Day())
		}
	}
	return
}

// GetUser get user by username or first and last name
func GetUser(username string) (user *tgbotapi.User, err error) {
	if len(username) == 0 {
		return
	}

	if useCouchbase {
		return GetUserCouchbase(username)
	} else if usePGSQL {
		return GetUserPGSQL(username)
	}

	return
}
