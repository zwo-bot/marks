package db

type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"column:name"`
}

type Favicon struct {
	ID     uint   `gorm:"primaryKey"`
	Data   []byte `gorm:"column:data"`
	Domain string `gorm:"column:domain"`
}

type Bookmark struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"column:title"`
	Path        string `gorm:"column:path"`
	Description string `gorm:"column:description"`
	URI         string `gorm:"column:uri"`
	Domain      string `gorm:"column:domain"`
	Tags        []Tag  `gorm:"many2many:bookmark_tags;"`
	Source      string `gorm:"column:source"`
}
