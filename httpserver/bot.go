package httpserver

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

	link, err := s.Bot.GetFileDirectURL(res.FileID)
	if err != nil {
		log.Printf(err.Error())
		return
	}
	filename := fmt.Sprintf("static/%d.jpg", chatID)
	go downloadImage(link, filename)

	result = filename

	return
}

func goDownloadImage(url, filename string) {
	err := downloadImage(url, filename)
	if err != nil {
		log.Printf("Error in download image: %s", err)
	}
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
