package entity

type EntityDaoMapImpl struct {
	// map attribute, entity
	stringAttributes map[string]map[string]string
	numberAttributes map[string]map[string]float64
	integerAttributes map[string]map[string]int64
}

func (dao EntityDaoMapImpl) GetEntity(id string) Entity {
	var e Entity
	e.id = id
	e.dao = dao
	return e;
}
	
func (dao EntityDaoMapImpl) SetEntity(entity Entity) {
	// Nothing to do
}

func (dao EntityDaoMapImpl) GetStringAttribute(entityId string, attributeId string) string {
	return dao.stringAttributes[attributeId][entityId]
}
func (dao EntityDaoMapImpl) GetNumberAttribute(entityId string, attributeId string) float64 {
	return dao.numberAttributes[attributeId][entityId]
}
func (dao EntityDaoMapImpl) GetIntegerAttribute(entityId string, attributeId string) int64 {
	return dao.integerAttributes[attributeId][entityId]
}

func (dao EntityDaoMapImpl) SetStringAttribute(entityId string, attributeId string, value string) {
	if (dao.stringAttributes[attributeId] == nil) {
		dao.stringAttributes[attributeId] = make(map[string]string)
	}
	dao.stringAttributes[attributeId][entityId] = value
}
func (dao EntityDaoMapImpl) SetIntegerAttribute(entityId string, attributeId string, value int64){
	if (dao.integerAttributes[attributeId] == nil) {
		dao.integerAttributes[attributeId] = make(map[string]int64)
	}
	dao.integerAttributes[attributeId][entityId] = value

}
func (dao EntityDaoMapImpl) SetNumberAttribute(entityId string, attributeId string, value float64) {
	if (dao.numberAttributes[attributeId] == nil) {
		dao.numberAttributes[attributeId] = make(map[string]float64)
	}
	dao.numberAttributes[attributeId][entityId] = value
}

func (dao EntityDaoMapImpl) RemoveStringAttribute(entityId string, attributeId string) {
	delete(dao.stringAttributes[attributeId],entityId)
}
func (dao EntityDaoMapImpl) RemoveNumberAttribute(entityId string, attributeId string) {
	delete(dao.numberAttributes[attributeId],entityId)
}
func (dao EntityDaoMapImpl) RemoveIntegerAttribute(entityId string, attributeId string) {
	delete(dao.integerAttributes[attributeId],entityId)
}
func (dao EntityDaoMapImpl) GetEntitiesWithStringAttribute(attribute string, value string) []string {
	attributeMap := make(map[string]string)
	attributeMap[attribute] = value
	return dao.GetEntitiesWithAttributes(attributeMap,nil,nil)
}
func (dao EntityDaoMapImpl) GetEntitiesWithNumberAttribute(attribute string, value float64) []string {
	attributeMap := make(map[string]float64)
	attributeMap[attribute] = value
	return dao.GetEntitiesWithAttributes(nil, attributeMap,nil)
}
func (dao EntityDaoMapImpl) GetEntitiesWithIntegerAttribute(attribute string, value int64) []string {
	attributeMap := make(map[string]int64)
	attributeMap[attribute] = value
	return dao.GetEntitiesWithAttributes(nil,nil,attributeMap)
}

func (dao EntityDaoMapImpl) GetEntitiesWithAttributes(strings map[string]string, numbers map[string]float64, ints map[string]int64) []string {
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

	var finalEntityList []string = nil
	for _, entities := range attributeEntities {
		if finalEntityList == nil {
			finalEntityList = entities
		} else {
			finalEntityList = intersection(entities,finalEntityList)
		}
	}
	return finalEntityList
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

