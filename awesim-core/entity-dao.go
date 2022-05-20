package main

type entityDao interface {

	getEntity(id string) entity
	
	setEntity(entity entity)

	getStringAttribute(entityId string, attributeId string) string
	getStringNumberAttribute(entityId string, attributeId string) float64
	getIntegerAttribute(entityId string, attributeId string) int64

	setStringAttribute(entityId string, attributeId string, value string)
	setIntegerAttribute(entityId string, attributeId string, value int64)
	setNumberAttribute(entityId string, attributeId string, value float64)

	getEntitiesWithAttributes(attributes []attribute) entity[]
	
}
