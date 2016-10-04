package httpserver

import (
	"gopkg.in/telegram-bot-api.v4"
)

func (s *Server) GetPhoto(chatID int64) (result int, err error) {
	config := tgbotapi.NewUserProfilePhotos(int(chatID))
	photos, err := s.Bot.GetUserProfilePhotos(config)
	if err != nil {
		return
	}

	result = photos.TotalCount
	return
}
