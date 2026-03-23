package storage

import "strings"
import "fmt"

import "github.com/patrickmn/go-cache"

type CachedStore struct {
	mainStore EntityDao
	cache cache.New()
}

func NewCachedStore(store EntityDao) *CachedStore {
	return &CachedStore{
		mainStore: store,
	}
}
	
func (dao *CachedStore) RemoveEntity(id string) error {
}

func (dao *CachedStore) GetAttribute(entityID string, attributeID string) (interface{}, error) {
	key := getKey(entityID, attributeID)
	val, ok := dao.attributeMap[key]
	if !ok {
		return nil, fmt.Errorf("Entity %s does not have attribute %s", entityID, attributeID)
	}
	return val, nil
}

func (dao *CachedStore) HasAttribute(entityID string, attributeID string) (bool, error) {
	key := getKey(entityID, attributeID)
	_, ok := dao.attributeMap[key]
	return ok, nil
}

func (dao *CachedStore) SetAttribute(entityID string, attributeID string, value interface{}) error {
	switch v := value.(type) {
	case string,int64,float64,bool:
		// It is a correct type
	default:
		return fmt.Errorf("Type %v not supported", v)
	}
	dao.attributeMap[getKey(entityID,attributeID)] = value
	return nil
}

func (dao *CachedStore) RemoveAttribute(entityID string, attributeID string) error {
	delete(dao.attributeMap, getKey(entityID,attributeID))
	return nil
}

func (dao *CachedStore) GetEntitiesWithAttributeType(attribute string) ([]string, error) {
	var entities []string
	for key, _ := range dao.attributeMap {
		if getAttributeFromKey(key) == attribute {
			entities = append(entities, getEntityIDFromKey(key))
		}
	}
	return entities, nil
}

func (dao *CachedStore) GetEntitiesWithAttribute(attribute string, value interface{}) ([]string, error) {
	attributeMap := make(map[string]interface{})
	attributeMap[attribute] = value
	return dao.GetEntitiesWithAttributes(attributeMap)
}

func (dao *CachedStore) GetEntitiesWithAttributes(attributes map[string]interface{}) ([]string, error) {
	countMap := make(map[string]int)
	for daoKey, daoValue := range dao.attributeMap {
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
	return fmt.Sprintf("%s#%s",entityID,attributeID)
}