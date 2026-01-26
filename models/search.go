package models

import "time"

type Search struct {
	Id        *string  `json:"id" validate:"required,uuid4"`
	Name      *string  `json:"name" validate:"required,min=3,max=100"`
	Image     *string  `json:"image" validate:"required,url"`
	ContentId *string  `json:"content_id" validate:"required,min=24,max=24"`
	Type      *string  `json:"type" validate:"required,eq=ARTIST|eq=ALBUM|eq=SONG|eq=PLAYLIST"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
}