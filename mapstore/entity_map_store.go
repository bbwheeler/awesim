package dao

import (
	"fmt"
	"strings"
)

type EntityMapStore struct {
	attributeMap map[string]any
}

func NewEntityMapStore() *EntityMapStore {
	attributes := make(map[string]any)
	return &EntityMapStore{
		attributeMap: attributes,
	}
}

func (entityMapStore *EntityMapStore) RemoveEntity(id string) error {
	for key := range entityMapStore.attributeMap {
		if getEntityIDFromKey(key) == id {
			delete(entityMapStore.attributeMap, key)
		}
	}
	return nil
}

func (entityMapStore *EntityMapStore) GetAttribute(entityID string, attributeID string) (any, error) {
	key := getKey(entityID, attributeID)
	val, ok := entityMapStore.attributeMap[key]
	if !ok {
		return nil, fmt.Errorf("Entity %s does not have attribute %s", entityID, attributeID)
	}
	return val, nil
}

func (entityMapStore *EntityMapStore) HasAttribute(entityID string, attributeID string) (bool, error) {
	key := getKey(entityID, attributeID)
	_, ok := entityMapStore.attributeMap[key]
	return ok, nil
}

func (entityMapStore *EntityMapStore) SetAttribute(entityID string, attributeID string, value any) error {
	switch v := value.(type) {
	case string, int64, float64, bool:
		// It is a correct type
	default:
		return fmt.Errorf("Type %v not supported", v)
	}
	entityMapStore.attributeMap[getKey(entityID, attributeID)] = value
	return nil
}

func (entityMapStore *EntityMapStore) RemoveAttribute(entityID string, attributeID string) error {
	delete(entityMapStore.attributeMap, getKey(entityID, attributeID))
	return nil
}

func (entityMapStore *EntityMapStore) GetEntitiesWithAttributeType(attribute string) ([]string, error) {
	var entities []string
	for key := range entityMapStore.attributeMap {
		if getAttributeFromKey(key) == attribute {
			entities = append(entities, getEntityIDFromKey(key))
		}
	}
	return entities, nil
}

func (entityMapStore *EntityMapStore) GetEntitiesWithAttribute(attribute string, value any) ([]string, error) {
	attributeMap := make(map[string]any)
	attributeMap[attribute] = value
	return entityMapStore.GetEntitiesWithAttributes(attributeMap)
}

func (entityMapStore *EntityMapStore) GetEntitiesWithAttributes(attributes map[string]any) ([]string, error) {
	countMap := make(map[string]int)
	for daoKey, daoValue := range entityMapStore.attributeMap {
		att := getAttributeFromKey(daoKey)
		ent := getEntityIDFromKey(daoKey)
		val, exists := attributes[att]
		if exists && val == daoValue {
			countMap[ent] = countMap[ent] + 1
		}
	}

	var ents []string
	for ent, num := range countMap {
		if num == len(attributes) {
			ents = append(ents, ent)
		}
	}
	return ents, nil
}

func intersection(list ...[]string) []string {
	if len(list) < 1 {
		return []string{}
	}

	theList := list[0]

	for _, loo := range list[1:] {
		for _, item := range theList {
			if !contains(loo, item) {
				i := indexOf(theList, item)
				if i >= 0 {
					theList = removeAtIndex(theList, i)
				}
			}
		}
	}
	return theList
}

func contains(s []string, val string) bool {
	return indexOf(s, val) >= 0
}
func indexOf(s []string, val string) int {
	for index, a := range s {
		if a == val {
			return index
		}
	}
	return -1
}

func removeAtIndex(s []string, index int) []string {
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}

func getEntityIDFromKey(key string) string {
	subs := strings.SplitN(key, "#", 2)
	return subs[0]

}
func getAttributeFromKey(key string) string {
	subs := strings.SplitN(key, "#", 2)
	return subs[1]
}
func getKey(entityID string, attributeID string) string {
	return fmt.Sprintf("%s#%s", entityID, attributeID)
}
