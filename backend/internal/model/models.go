package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         string         `gorm:"size:64;not null" json:"name"`
	Email        string         `gorm:"size:128;uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Role         string         `gorm:"size:16;not null;default:user" json:"role"`
	AvatarURL    string         `gorm:"size:500" json:"avatarUrl"`
}

type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `gorm:"size:64;uniqueIndex;not null" json:"name"`
	Slug      string    `gorm:"size:80;uniqueIndex;not null" json:"slug"`
	Articles  []Article `json:"-"`
}

type Article struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	Title          string     `gorm:"size:180;not null" json:"title"`
	Slug           string     `gorm:"size:220;uniqueIndex;not null" json:"slug"`
	Summary        string     `gorm:"type:text" json:"summary"`
	CoverURL       string     `gorm:"size:500" json:"coverUrl"`
	Content        string     `gorm:"type:longtext;not null" json:"content"`
	Tags           string     `gorm:"type:text" json:"tags"`
	Status         string     `gorm:"size:16;index;not null;default:draft" json:"status"`
	PublishedAt    *time.Time `json:"publishedAt"`
	CategoryID     *uint      `gorm:"index" json:"categoryId"`
	Category       Category   `json:"category"`
	AuthorID       uint       `gorm:"index" json:"authorId"`
	Author         User       `json:"author"`
	LikesCount     int64      `gorm:"not null;default:0" json:"likesCount"`
	FavoritesCount int64      `gorm:"not null;default:0" json:"favoritesCount"`
	Comments       []Comment  `json:"comments,omitempty"`
}

type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	ArticleID uint      `gorm:"index;not null" json:"articleId"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	Status    string    `gorm:"size:16;index;not null;default:published" json:"status"`
	User      User      `json:"user"`
	Article   Article   `json:"article"`
}

type Reaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uint      `gorm:"uniqueIndex:ux_user_article_type,priority:1;not null" json:"userId"`
	ArticleID uint      `gorm:"uniqueIndex:ux_user_article_type,priority:2;not null;index" json:"articleId"`
	Type      string    `gorm:"uniqueIndex:ux_user_article_type,priority:3;size:16;not null" json:"type"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Type      string    `gorm:"size:24;index;not null" json:"type"`
	Title     string    `gorm:"size:180;not null" json:"title"`
	Content   string    `gorm:"size:500" json:"content"`
	IsRead    bool      `gorm:"index;not null;default:false" json:"isRead"`
	UserID    *uint     `gorm:"index" json:"userId"`
	ArticleID *uint     `gorm:"index" json:"articleId"`
	CommentID *uint     `gorm:"index" json:"commentId"`
	User      User      `json:"user,omitempty"`
	Article   Article   `json:"article,omitempty"`
	Comment   Comment   `json:"comment,omitempty"`
}
