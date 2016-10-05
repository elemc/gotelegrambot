package httpserver

import (
	"RussianFedoraBot/db"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"gopkg.in/telegram-bot-api.v4"
)

// PhotosCache type for store users photo filenames by id
type PhotosCache map[int64]string

// UpdatePhotoCache function update photos cache of users
func (s *Server) UpdatePhotoCache() {
	users, err := db.GetUsers()
	if err != nil {
		log.Printf("Error in UpdatePhotoCache: %s", err)
		return
	}

	s.PhotoCache = make(PhotosCache) // new cache

	for _, user := range users {
		go s.GetPhoto(int64(user.ID))
	}
}

// GetPhotoFileName returns name photo file
func (s *Server) GetPhotoFileName(userID int64) (result string) {
	if fn, ok := s.PhotoCache[userID]; ok {
		result = getFileName(fn)
	} else {
		result = getFileName("nobody.png")
	}
	return
}

// GetPhoto fucntion download user photo and return file name for html tag img
func (s *Server) GetPhoto(chatID int64) {
	config := tgbotapi.NewUserProfilePhotos(int(chatID))
	photos, err := s.Bot.GetUserProfilePhotos(config)
	if err != nil {
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
	err = downloadImage(link, getFileName(filename))
	if err != nil {
		log.Printf("Error in downloadImage: %s", err)
		return
	}
	s.PhotoCache[chatID] = filename

	return
}

func getFileName(fn string) string {
	return fmt.Sprintf("static/%s", fn)
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
