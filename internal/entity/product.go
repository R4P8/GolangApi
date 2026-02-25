package entity

import "time"

type Product struct {
	IDProduct   int       `json:"id_product"`
	IDCategory  int       `json:"id_category"`
	Name        string    `json:"name"`
	Qty         int       `json:"qty"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
