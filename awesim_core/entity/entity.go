package entity

const TypeAttribute string = "TYPE"


type Entity struct {

	id string

	dao EntityDao

}

func (e Entity) GetStringAttribute(attributeId string) string {
	return e.dao.GetStringAttribute(e.id, attributeId)
}
func (e Entity) GetNumberAttribute(attributeId string) float64 {
	return e.dao.GetNumberAttribute(e.id, attributeId)
}
func (e Entity) GetIntegerAttribute(attributeId string) int64 {
	return e.dao.GetIntegerAttribute(e.id, attributeId)
}

func (e Entity) SetStringAttribute(attributeId string, value string) {
	e.dao.SetStringAttribute(e.id, attributeId, value);
}
func (e Entity) SetNumberAttribute(attributeId string, value float64) {
	e.dao.SetNumberAttribute(e.id, attributeId, value);
}
func (e Entity) SetIntegerAttribute(attributeId string, value int64) {
	e.dao.SetIntegerAttribute(e.id, attributeId, value);
}
