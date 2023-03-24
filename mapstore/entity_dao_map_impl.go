package dao

import "github.com/bbwheeler/awesim/core"

import "strings"
import "fmt"

type EntityDaoMapImpl struct {
	// map entity#attribute, value
	attributeMap map[string]interface{}
}

func NewEntityDaoMapImpl() *EntityDaoMapImpl {
	attributes := make(map[string]interface{})
	return &EntityDaoMapImpl{
		attributeMap: attributes,
	}
}

func (dao *EntityDaoMapImpl) NewEntity() *core.Entity {
	return core.NewEntity(dao)
}

func (dao *EntityDaoMapImpl) GetEntity(id string) *core.Entity {
	return core.GetEntity(id, dao)
}
	
func (dao *EntityDaoMapImpl) RemoveEntity(id string) error {
	for key, _ := range dao.attributeMap {
		if getEntityIDFromKey(key) == id {
			delete(dao.attributeMap, key)
		}
	}
	return nil
}

func (dao *EntityDaoMapImpl) GetAttribute(entityID string, attributeID string) (interface{}, error) {
	key := getKey(entityID, attributeID)
	val, ok := dao.attributeMap[key]
	if !ok {
		return nil, fmt.Errorf("Entity %s does not have attribute %s", entityID, attributeID)
	}
	return val, nil
}

func (dao *EntityDaoMapImpl) SetAttribute(entityID string, attributeID string, value interface{}) error {
	switch v := value.(type) {
	case string,int64,float64,bool:
		// It is a correct type
	default:
		return fmt.Errorf("Type %v not supported", v)
	}
	dao.attributeMap[getKey(entityID,attributeID)] = value
	return nil
}

func (dao *EntityDaoMapImpl) RemoveAttribute(entityID string, attributeID string) error {
	delete(dao.attributeMap, getKey(entityID,attributeID))
	return nil
}

func (dao *EntityDaoMapImpl) GetEntitiesWithAttributeType(attribute string) ([]*core.Entity, error) {
	var entities []*core.Entity
	for key, _ := range dao.attributeMap {
		if getAttributeFromKey(key) == attribute {
			entities = append(entities, dao.GetEntity(getEntityIDFromKey(key)))
		}
	}
	return entities, nil
}

func (dao *EntityDaoMapImpl) GetEntitiesWithAttribute(attribute string, value interface{}) ([]*core.Entity, error) {
	attributeMap := make(map[string]interface{})
	attributeMap[attribute] = value
	return dao.GetEntitiesWithAttributes(attributeMap)
}

func (dao *EntityDaoMapImpl) GetEntitiesWithAttributes(attributes map[string]interface{}) ([]*core.Entity, error) {
	var entitiesForEachAttribute [][]string

	// This can definitely be more efficient than it is
	for _, attributeValue := range attributes {
		var entityList []string
		for key, val := range dao.attributeMap {
			attKey := getAttributeFromKey(key)
			if attKey == key && val == attributeValue {
				 entityList = append(entityList, getEntityIDFromKey(key))
			}
		}		
		entitiesForEachAttribute = append(entitiesForEachAttribute, entityList)
	}


	var finalEntityList []string = nil
	for _, entities := range entitiesForEachAttribute {
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