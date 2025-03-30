package models

type PrivilegeCategory struct {
	Category    string ` rethink:"category" `
	Description string ` rethink:"decription" `
}
