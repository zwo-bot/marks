package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/zwo-bot/go-rofi-bookmarks/bookmark"
)

var DB *gorm.DB

func ConnectDatabase() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("bookmarks.db"), &gorm.Config{})
	if err == nil {
		err = DB.AutoMigrate(&Tag{}, &Favicon{}, &Bookmark{})
	}
	return err
}

func CloseDatabase() error {
	db, err := DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func GetBookmarks() (bookmark.Bookmarks, error) {
	var bookmarks bookmark.Bookmarks
	err := DB.Preload("Tags").Find(&bookmarks).Error
	return bookmarks, err
}

func SaveBookmark(bookmark bookmark.Bookmark) error {
	return DB.Save(&bookmark).Error
}