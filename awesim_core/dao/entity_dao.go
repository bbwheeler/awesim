package dao

import "github.com/bbwheeler/awesim/awesim_core/core"

type EntityDao interface {

	GetEntity(id string) core.Entity
	SetEntity(entity core.Entity)

	GetStringAttribute(entityId string, attributeId string) (string, error)
	GetNumberAttribute(entityId string, attributeId string) (float64, error)
	GetIntegerAttribute(entityId string, attributeId string) (int64, error)
	GetBooleanAttribute(entityId string, attributeId string) (bool, error)

	SetStringAttribute(entityId string, attributeId string, value string) error
	SetIntegerAttribute(entityId string, attributeId string, value int64) error
	SetNumberAttribute(entityId string, attributeId string, value float64) error
	SetBooleanAttribute(entityId string, attributeId string, value bool) error

	RemoveStringAttribute(entityId string, attributeId string) error
	RemoveIntegerAttribute(entityId string, attributeId string) error
	RemoveNumberAttribute(entityId string, attributeId string) error
	RemoveBooleanAttribute(entityId string, attributeId string) error

	GetEntitiesWithAttributes(strings map[string]string, numbers map[string]float64, ints map[string]int64, bools map[string]bool) ([]core.Entity, error)
	GetEntitiesWithStringAttribute(attribute string, value string) ([]core.Entity, error)
	GetEntitiesWithIntegerAttribute(attribute string, value int64) ([]core.Entity, error)
	GetEntitiesWithNumberAttribute(attribute string, value float64) ([]core.Entity, error)
	GetEntitiesWithBooleanAttribute(attribute string, value bool) ([]core.Entity, error)
	
}
