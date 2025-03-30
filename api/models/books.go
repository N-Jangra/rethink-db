package models

import "time"

type Books struct {
	BookID      int       ` json:"bookid,omitempty" rethink:"bookid" `
	Title       string    ` json:"title" rethink:"title" `
	Description string    ` json:"description" rethink:"description" `
	CreatedBy   string    ` json:"createdby" rethink:"createdby" `
	CreatedAt   time.Time ` json:"createdat" rethink:"createdat" `
	UpdatedBy   string    ` json:"updatedby" rethink:"updatedby" `
	UpdatedAt   time.Time ` json:"updatedat" rethink:"updatedat" `
}

func (Books) TableName() string {
	return "books"
}
