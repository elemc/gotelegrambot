package httpserver

import (
	"log"

	"gopkg.in/telegram-bot-api.v4"
)

func (s *Server) GetPhoto(chatID int64) (result string, err error) {
	config := tgbotapi.NewUserProfilePhotos(int(chatID))
	photos, err := s.Bot.GetUserProfilePhotos(config)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	res := photos.Photos[0][0]

	result, err = s.Bot.GetFileDirectURL(res.FileID)

	return
}
