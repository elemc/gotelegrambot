package main

import (
	"RussianFedoraBot/db"
	"RussianFedoraBot/httpserver"
	"log"

	"gopkg.in/telegram-bot-api.v4"
)

var bot *tgbotapi.BotAPI

func main() {
	bot, err := tgbotapi.NewBotAPI("290179858:AAFvx-ekOd7OkPkQYnGVggakR12BemcpxVI")
	if err != nil {
		log.Fatalf("Cannot create bot api")
	}

	bot.Debug = false // true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// start http server
	s := httpserver.Server{Addr: ":8088", Bot: bot}
	s.PhotoCache = make(httpserver.PhotosCache)
	go s.Start()
	//s.Start()

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
		id := int64(update.Message.From.ID)
		if _, ok := s.PhotoCache[id]; !ok {
			go s.GetPhoto(id)
		}
	}
}
