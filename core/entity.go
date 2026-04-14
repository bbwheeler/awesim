package core

import "github.com/google/uuid"

type errString string

func (e errString) Error() string {
	return string(e)
}

type Attribute interface {
	string | int64 | float64 | bool
}

type Entity struct {
	id    string
	store EntityStore
}

type EntityStore interface {
	GetAttribute(entityId string, attributeId string) (interface{}, error)
	HasAttribute(entityId string, attributeId string) (bool, error)
	SetAttribute(entityId string, attributeId string, value interface{}) error
	RemoveAttribute(entityId string, attributeId string) error
	GetEntitiesWithAttribute(attributeID string, attributeValue any) ([]string, error)
	GetEntitiesWithAttributeType(attribute string) ([]string, error)
}

func NewEntity(store EntityStore) *Entity {
	return GetEntity(uuid.New().String(), store)

}

func GetEntity(id string, store EntityStore) *Entity {
	return &Entity{
		id:    id,
		store: store,
	}
}

func (e *Entity) GetID() string {
	return e.id
}

func (e *Entity) GetAttribute(attribute string) (any, error) {
	return e.store.GetAttribute(e.GetID(), attribute)
}
func (e *Entity) SetAttribute(attribute string, value any) error {
	return e.store.SetAttribute(e.GetID(), attribute, value)
}
func (e *Entity) RemoveAttribute(attribute string) error {
	return e.store.RemoveAttribute(e.GetID(), attribute)
}
func (e *Entity) HasAttribute(attribute string) (bool, error) {
	return e.store.HasAttribute(e.GetID(), attribute)
}
