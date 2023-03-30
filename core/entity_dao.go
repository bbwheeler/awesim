package core

type EntityDao interface {

	RemoveEntity(entityID string) error

	GetAttribute(entityId string, attributeId string) (interface{}, error)
	SetAttribute(entityId string, attributeId string, value interface{}) error
	RemoveAttribute(entityId string, attributeId string) error

	GetEntitiesWithAttributes(attributes map[string]interface{}) ([]string, error)
	GetEntitiesWithAttributeType(attribute string) ([]string, error)
	GetEntitiesWithAttribute(attribute string, value interface{}) ([]string, error)	
}
