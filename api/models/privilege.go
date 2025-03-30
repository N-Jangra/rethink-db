package models

type Privilege struct {
	Privilege   string ` rethink:"privilege" `
	Category    string ` rethink:"category" `
	Description string ` rethink:"decription" `
	Type        string ` rethink:"type" `
	AppId       string ` rethink:"appid" `
}
