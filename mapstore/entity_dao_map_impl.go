package dao

import "github.com/bbwheeler/awesim/core"


type EntityDaoMapImpl struct {
	// map attribute, entity
	attributes map[string]map[string]interface{}
}

func NewEntityDaoMapImpl() *EntityDaoMapImpl {
	attributes := make(map[string]map[string]interface{})
	return &EntityDaoMapImpl{
		attributes: attributes,
	}
}

func (dao *EntityDaoMapImpl) NewEntity() *core.Entity {
	return core.NewEntity(dao)
}

func (dao *EntityDaoMapImpl) GetEntity(id string) *core.Entity {
	return core.GetEntity(id, dao)
}
	
func (dao *EntityDaoMapImpl) RemoveEntity(entity *core.Entity) error {
	entityID := entity.GetID()
	for attributeKey, _ := range dao.attributes {
		entityMap := dao.attributes[attributeKey]
		delete(entityMap, entityID)
	}
	return nil
}

func (dao *EntityDaoMapImpl) GetAttribute(entityId string, attributeId string) (interface{}, error) {
	return dao.attributes[attributeId][entityId], nil
}

func (dao *EntityDaoMapImpl) SetAttribute(entityId string, attributeId string, value interface{}) error {
	if (dao.attributes[attributeId] == nil) {
		dao.attributes[attributeId] = make(map[string]interface{})
	}
	dao.attributes[attributeId][entityId] = value
	return nil
}

func (dao *EntityDaoMapImpl) RemoveAttribute(entityId string, attributeId string) error {
	delete(dao.attributes[attributeId],entityId)
	return nil
}

func (dao *EntityDaoMapImpl) GetEntitiesWithAttributeType(attribute string) ([]*core.Entity, error) {
	var attributeEntities []string
	for attributeKey, _ := range dao.attributes {
		entityMap := dao.attributes[attributeKey]
		for entityKey, _ := range entityMap {
			attributeEntities = append(attributeEntities, entityKey)
		}
	}

	var finalEntities []*core.Entity
	for _, entityID := range attributeEntities {
		finalEntities = append(finalEntities, dao.GetEntity(entityID))
	}

	return finalEntities, nil
}

func (dao *EntityDaoMapImpl) GetEntitiesWithAttribute(attribute string, value interface{}) ([]*core.Entity, error) {
	attributeMap := make(map[string]interface{})
	attributeMap[attribute] = value
	return dao.GetEntitiesWithAttributes(attributeMap)
}

func (dao *EntityDaoMapImpl) GetEntitiesWithAttributes(attributes map[string]interface{}) ([]*core.Entity, error) {
	attributeEntities := make(map[string][]string)
	for attributeKey, attributeValue := range attributes {
		entityMap := dao.attributes[attributeKey]
		for entityKey, entityAttributeValue := range entityMap {
			if attributeValue == entityAttributeValue {
				if attributeEntities[attributeKey] == nil {
					attributeEntities[attributeKey] = []string{}
				}
				attributeEntities[attributeKey] = append(attributeEntities[attributeKey],entityKey)
			} 
		}
	}

	var finalEntityList []string = nil
	for _, entities := range attributeEntities {
		if finalEntityList == nil {
			finalEntityList = entities
		} else {
			finalEntityList = intersection(entities,finalEntityList)
		}
	}

	var finalEntities []*core.Entity
	for _, entityID := range finalEntityList {
		finalEntities = append(finalEntities, dao.GetEntity(entityID))
	}

	return finalEntities, nil
}

func intersection(a, b []string) (c []string) {
	m := make(map[string]bool)

	for _, item := range a {
			m[item] = true
	}

	for _, item := range b {
			if _, ok := m[item]; ok {
					c = append(c, item)
			}
	}
	return
}

