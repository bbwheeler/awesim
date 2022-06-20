package entity

type EntityDao interface {

	GetEntity(id string) Entity
	SetEntity(entity Entity)

	GetStringAttribute(entityId string, attributeId string) string
	GetNumberAttribute(entityId string, attributeId string) float64
	GetIntegerAttribute(entityId string, attributeId string) int64

	SetStringAttribute(entityId string, attributeId string, value string)
	SetIntegerAttribute(entityId string, attributeId string, value int64)
	SetNumberAttribute(entityId string, attributeId string, value float64)

	RemoveStringAttribute(entityId string, attributeId string)
	RemoveIntegerAttribute(entityId string, attributeId string)
	RemoveNumberAttribute(entityId string, attributeId string)

	GetEntitiesWithAttributes(strings map[string]string, numbers map[string]float64, ints map[string]int64) []string
	GetEntitiesWithStringAttribute(attribute string, value string) []string
	GetEntitiesWithIntegerAttribute(attribute string, value int64) []string
	GetEntitiesWithNumberAttribute(attribute string, value float64) []string
	
}
