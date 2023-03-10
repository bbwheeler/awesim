package dao

import "github.com/bbwheeler/awesim/awesim_core/core"

type EntityDaoMapImpl struct {
	// map attribute, entity
	stringAttributes map[string]map[string]string
	numberAttributes map[string]map[string]float64
	integerAttributes map[string]map[string]int64
	booleanAttributes map[string]map[string]bool
}

func NewEntityDaoMapImpl() *EntityDaoMapImpl {
	stringAttributes := make(map[string]map[string]string)
	numberAttributes := make(map[string]map[string]float64)
	integerAttributes := make(map[string]map[string]int64)
	booleanAttributes := make(map[string]map[string]bool)
	return &EntityDaoMapImpl{
		stringAttributes: stringAttributes,
		numberAttributes: numberAttributes,
		integerAttributes: integerAttributes,
		booleanAttributes: booleanAttributes,
	}
}


type entity struct {
	id string
	dao EntityDao
}

func (e entity) GetID() string {
	return e.id	
}

func (e entity) GetStringAttribute(attribute string) (string, error) {
	return e.dao.GetStringAttribute(e.GetID(), attribute)
}
func (e entity) GetNumberAttribute(attribute string) (float64, error) {
	return e.dao.GetNumberAttribute(e.GetID(), attribute)
}
func (e entity) GetIntegerAttribute(attribute string) (int64, error) {
	return e.dao.GetIntegerAttribute(e.GetID(), attribute)
}
func (e entity) GetBooleanAttribute(attribute string) (bool, error) {
	return e.dao.GetBooleanAttribute(e.GetID(), attribute)
}

func (dao *EntityDaoMapImpl) GetEntity(id string) core.Entity {
	return entity{
		id: id,
		dao: dao,
	}
}
	
func (dao *EntityDaoMapImpl) SetEntity(entity core.Entity) {
	// Nothing to do
}

func (dao *EntityDaoMapImpl) GetStringAttribute(entityId string, attributeId string) (string, error) {
	return dao.stringAttributes[attributeId][entityId], nil
}
func (dao *EntityDaoMapImpl) GetNumberAttribute(entityId string, attributeId string) (float64, error) {
	return dao.numberAttributes[attributeId][entityId], nil
}
func (dao *EntityDaoMapImpl) GetIntegerAttribute(entityId string, attributeId string) (int64, error) {
	return dao.integerAttributes[attributeId][entityId], nil
}
func (dao *EntityDaoMapImpl) GetBooleanAttribute(entityId string, attributeId string) (bool, error) {
	return dao.booleanAttributes[attributeId][entityId], nil
}

func (dao *EntityDaoMapImpl) SetStringAttribute(entityId string, attributeId string, value string) error {
	if (dao.stringAttributes[attributeId] == nil) {
		dao.stringAttributes[attributeId] = make(map[string]string)
	}
	dao.stringAttributes[attributeId][entityId] = value
	return nil
}
func (dao *EntityDaoMapImpl) SetIntegerAttribute(entityId string, attributeId string, value int64) error {
	if (dao.integerAttributes[attributeId] == nil) {
		dao.integerAttributes[attributeId] = make(map[string]int64)
	}
	dao.integerAttributes[attributeId][entityId] = value
	return nil
}
func (dao *EntityDaoMapImpl) SetNumberAttribute(entityId string, attributeId string, value float64) error {
	if (dao.numberAttributes[attributeId] == nil) {
		dao.numberAttributes[attributeId] = make(map[string]float64)
	}
	dao.numberAttributes[attributeId][entityId] = value
	return nil
}
func (dao *EntityDaoMapImpl) SetBooleanAttribute(entityId string, attributeId string, value bool) error {
	if (dao.booleanAttributes[attributeId] == nil) {
		dao.booleanAttributes[attributeId] = make(map[string]bool)
	}
	dao.booleanAttributes[attributeId][entityId] = value
	return nil
}

func (dao *EntityDaoMapImpl) RemoveStringAttribute(entityId string, attributeId string) error {
	delete(dao.stringAttributes[attributeId],entityId)
	return nil
}
func (dao *EntityDaoMapImpl) RemoveNumberAttribute(entityId string, attributeId string) error {
	delete(dao.numberAttributes[attributeId],entityId)
	return nil
}
func (dao *EntityDaoMapImpl) RemoveIntegerAttribute(entityId string, attributeId string) error {
	delete(dao.integerAttributes[attributeId],entityId)
	return nil
}
func (dao *EntityDaoMapImpl) RemoveBooleanAttribute(entityId string, attributeId string) error {
	delete(dao.booleanAttributes[attributeId],entityId)
	return nil
}

func (dao *EntityDaoMapImpl) GetEntitiesWithStringAttribute(attribute string, value string) ([]core.Entity, error) {
	attributeMap := make(map[string]string)
	attributeMap[attribute] = value
	return dao.GetEntitiesWithAttributes(attributeMap,nil,nil,nil)
}
func (dao *EntityDaoMapImpl) GetEntitiesWithNumberAttribute(attribute string, value float64) ([]core.Entity, error) {
	attributeMap := make(map[string]float64)
	attributeMap[attribute] = value
	return dao.GetEntitiesWithAttributes(nil, attributeMap,nil,nil)
}
func (dao *EntityDaoMapImpl) GetEntitiesWithIntegerAttribute(attribute string, value int64) ([]core.Entity, error) {
	attributeMap := make(map[string]int64)
	attributeMap[attribute] = value
	return dao.GetEntitiesWithAttributes(nil,nil,attributeMap,nil)
}
func (dao *EntityDaoMapImpl) GetEntitiesWithBooleanAttribute(attribute string, value bool) ([]core.Entity, error) {
	attributeMap := make(map[string]bool)
	attributeMap[attribute] = value
	return dao.GetEntitiesWithAttributes(nil,nil,nil,attributeMap)
}

func (dao *EntityDaoMapImpl) GetEntitiesWithAttributes(strings map[string]string, numbers map[string]float64, ints map[string]int64, bools map[string]bool) ([]core.Entity, error) {
	attributeEntities := make(map[string][]string)
	for attributeKey, attributeValue := range strings {
		entityMap := dao.stringAttributes[attributeKey]
		for entityKey, entityAttributeValue := range entityMap {
			if attributeValue == entityAttributeValue {
				if attributeEntities[attributeKey] == nil {
					attributeEntities[attributeKey] = []string{}
				}
				attributeEntities[attributeKey] = append(attributeEntities[attributeKey],entityKey)
			} 
		}
	}
	for attributeKey, attributeValue := range numbers {
		entityMap := dao.numberAttributes[attributeKey]
		for entityKey, entityAttributeValue := range entityMap {
			if attributeValue == entityAttributeValue {
				if attributeEntities[attributeKey] == nil {
					attributeEntities[attributeKey] = []string{}
				}
				attributeEntities[attributeKey] = append(attributeEntities[attributeKey],entityKey)
			} 
		}
	}
	for attributeKey, attributeValue := range ints {
		entityMap := dao.integerAttributes[attributeKey]
		for entityKey, entityAttributeValue := range entityMap {
			if attributeValue == entityAttributeValue {
				if attributeEntities[attributeKey] == nil {
					attributeEntities[attributeKey] = []string{}
				}
				attributeEntities[attributeKey] = append(attributeEntities[attributeKey],entityKey)
			} 
		}
	}
	for attributeKey, attributeValue := range bools {
		entityMap := dao.booleanAttributes[attributeKey]
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

	var finalEntities []core.Entity
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

