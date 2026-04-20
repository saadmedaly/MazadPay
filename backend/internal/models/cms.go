package models

import (
	"time"
)

type FAQItem struct {
	ID           int     `db:"id"             json:"id"`
	QuestionAr   string  `db:"question_ar"    json:"question_ar"`
	QuestionFr   *string `db:"question_fr"    json:"question_fr"`
	AnswerAr     string  `db:"answer_ar"      json:"answer_ar"`
	AnswerFr     *string `db:"answer_fr"      json:"answer_fr"`
	DisplayOrder int     `db:"display_order"  json:"display_order"`
}

type Tutorial struct {
	ID           int     `db:"id"             json:"id"`
	TitleAr      string  `db:"title_ar"       json:"title_ar"`
	TitleFr      *string `db:"title_fr"       json:"title_fr"`
	VideoURL     string  `db:"video_url"      json:"video_url"`
	ThumbnailURL *string `db:"thumbnail_url"  json:"thumbnail_url"`
	Category     *string `db:"category"       json:"category"`
	DisplayOrder int     `db:"display_order"  json:"display_order"`
}

type Banner struct {
	ID           int       `db:"id"             json:"id"`
	TitleAr      string    `db:"title_ar"       json:"title_ar"`
	TitleFr      string    `db:"title_fr"       json:"title_fr"`
	TitleEn      string    `db:"title_en"       json:"title_en"`
	ImageURL     string    `db:"image_url"      json:"image_url"`
	TargetURL   string    `db:"target_url"     json:"target_url"`
	IsActive     bool      `db:"is_active"      json:"is_active"`
	StartsAt     *time.Time `db:"starts_at"      json:"starts_at"`
	EndsAt       *time.Time `db:"ends_at"        json:"ends_at"`
	DisplayOrder int       `db:"display_order"  json:"display_order"`
}
