package httpserver

import (
	"RussianFedoraBot/db"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/telegram-bot-api.v4"
)

// PhotosCache type for store users photo filenames by id
type PhotosCache map[int64]string

// FilesCache type for store files
type FilesCache map[string]string

// UpdatePhotoCache function update photos cache of users
func (s *Server) UpdatePhotoCache() {
	users, err := db.GetUsers()
	if err != nil {
		log.Printf("Error in UpdatePhotoCache: %s", err)
		return
	}

	s.PhotoCache = make(PhotosCache) // new cache

	for _, user := range users {
		s.GetPhoto(int64(user.ID))
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

// GetFileNameByFileID returns file name by index
func (s *Server) GetFileNameByFileID(chatID int64, fileID string) (filename string) {
	f, err := db.GetFile(fileID, chatID)
	if err != nil {
		// try to download it
		s.GetFile(fileID, chatID)
		f, err = db.GetFile(fileID, chatID)
		if err != nil {
			log.Printf("Error in GetFileNameByFileID with FileID [%s]: %s", fileID, err)
			return "missing-data"
		}
	}
	filename = fmt.Sprintf("static/%s", f.FilePath)

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

// GetFile function for get file from telegram
func (s *Server) GetFile(fileID string, chatID int64) {
	fc := tgbotapi.FileConfig{}
	fc.FileID = fileID
	f, err := s.Bot.GetFile(fc)
	if err != nil {
		log.Printf("Error in GetFile for FileID [%s]: %s", fileID, err)
		return
	}

	log.Printf(f.FilePath)

	// check directory
	dir := filepath.Dir(f.FilePath)
	path := filepath.Join("static", dir)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		log.Printf("Error in MkdirAll for FileID [%s]: %s", fileID, err)
		return
	}

	filename := filepath.Join("static", f.FilePath)
	err = downloadImage(f.Link(s.APIKey), filename)
	if err != nil {
		log.Printf("Error in MkdirAll for FileID [%s]: %s", fileID, err)
		return
	}
	//s.FileCache[f.FileID] = filepath.Join("static", f.FilePath)
	err = db.SaveFile(&f, chatID)
	if err != nil {
		log.Printf("Error in SaveFile for FileID [%s]: %s", fileID, err)
	}
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
