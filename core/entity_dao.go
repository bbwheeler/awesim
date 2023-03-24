package core

type EntityDao interface {

	NewEntity() *Entity
	GetEntity(id string) *Entity
	RemoveEntity(entityID string) error

	GetAttribute(entityId string, attributeId string) (interface{}, error)
	SetAttribute(entityId string, attributeId string, value interface{}) error
	RemoveAttribute(entityId string, attributeId string) error

	GetEntitiesWithAttributes(attributes map[string]interface{}) ([]*Entity, error)
	GetEntitiesWithAttributeType(attribute string) ([]*Entity, error)
	GetEntitiesWithAttribute(attribute string, value interface{}) ([]*Entity, error)	
}
