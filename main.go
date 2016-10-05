package main

import (
	"RussianFedoraBot/db"
	"RussianFedoraBot/httpserver"
	"flag"
	"log"

	"gopkg.in/telegram-bot-api.v4"
)

var (
	bot      *tgbotapi.BotAPI
	settings Settings
)

func init() {
	LoadConfig()

	flag.StringVar(&settings.APIKey, "api-key", settings.APIKey, "API key for Telegram bot")
	flag.StringVar(&settings.Addr, "addr", settings.Addr, "address string host:port for listen http server")
	flag.StringVar(&settings.Couchbase.Cluster, "couch-cluster", settings.Couchbase.Cluster, "url to couchbase cluster")
	flag.StringVar(&settings.Couchbase.Bucket, "couch-bucket", settings.Couchbase.Bucket, "couchbase bucket name")
	flag.StringVar(&settings.Couchbase.Secret, "couch-secret", settings.Couchbase.Secret, "couchbase bucket password")
}

func main() {
	flag.Parse()
	//SaveConfig()
	db.InitCouchbase(settings.Couchbase.Cluster, settings.Couchbase.Bucket, settings.Couchbase.Secret)

	bot, err := tgbotapi.NewBotAPI(settings.APIKey)
	if err != nil {
		log.Fatalf("Cannot create bot api")
	}

	bot.Debug = false // true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// start http server
	s := httpserver.Server{Addr: settings.Addr, Bot: bot}
	s.PhotoCache = make(httpserver.PhotosCache)
	s.FileCache = make(httpserver.FilesCache)
	s.APIKey = settings.APIKey
	//go s.Start()
	s.Start()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err.Error())
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		go db.GoSaveMessage(update.Message)

		// Photo
		id := int64(update.Message.From.ID)
		if _, ok := s.PhotoCache[id]; !ok {
			go s.GetPhoto(id)
		}

		// Files
		if update.Message.Audio != nil {
			go s.GetFile(update.Message.Audio.FileID, update.Message.Chat.ID)
		}
		if update.Message.Document != nil {
			go s.GetFile(update.Message.Document.FileID, update.Message.Chat.ID)
		}
		if update.Message.Photo != nil {
			for _, f := range *update.Message.Photo {
				go s.GetFile(f.FileID, update.Message.Chat.ID)
			}
		}
		if update.Message.Sticker != nil {
			go s.GetFile(update.Message.Sticker.FileID, update.Message.Chat.ID)
		}
		if update.Message.Video != nil {
			go s.GetFile(update.Message.Video.FileID, update.Message.Chat.ID)
		}
		if update.Message.Voice != nil {
			go s.GetFile(update.Message.Voice.FileID, update.Message.Chat.ID)
		}
	}
}
