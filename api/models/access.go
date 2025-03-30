package models

type Access struct {
	Privilege string ` rethink:"privileges" `
	Role      string ` rethink:"role" `
}

func (Access) TableName() string {
	return "access"
}
