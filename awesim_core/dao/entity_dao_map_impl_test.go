package dao_test

import "testing"
import (
	"github.com/bbwheeler/awesim/awesim_core/entity"
)

func TestEntityDaoMapImpl(t *testing.T) {
	const mockEntityOne string = "entityOne"
	const mockAttributeOne string = "mock_attribute_one"
	const valueOne string = "a value"

	testDao := NewEntityDaoMapImpl()
	testDao.SetStringAttribute(mockEntityOne,mockAttributeOne,valueOne)
	var result string = testDao.GetStringAttribute(mockEntityOne,mockAttributeOne)
	if (result != valueOne) {
		t.Error("Attribute value should be 'a value' but was ", result)
	}
}