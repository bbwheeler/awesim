package core

import "github.com/google/uuid"

type Entity struct {
	id string
	dao EntityDao
}

func NewEntity(dao EntityDao) *Entity {
	return GetEntity(uuid.New().String(), dao)
	
}

func GetEntity(id string, dao EntityDao) *Entity {
	return &Entity{
		id: id,
		dao: dao,
	}
}

func (e *Entity) GetID() string {
	return e.id	
}

func (e *Entity) GetAttribute(attribute string) (interface{}, error) {
	return e.dao.GetAttribute(e.GetID(), attribute)
}
func (e *Entity) SetAttribute(attribute string, value interface{}) error {
	return e.dao.SetAttribute(e.GetID(), attribute, value)
}
func (e *Entity) RemoveAttribute(attribute string) error {
	return e.dao.RemoveAttribute(e.GetID(), attribute)
}
