package models

type Roles struct {
	Role        string ` rethink:"role" `
	Level       int16  ` rethink:"level" `
	Description string ` rethink:"decription" `
}
