package dao_test

import "testing"
import mapstore "github.com/bbwheeler/awesim/mapstore"



func EndToEndTest(t *testing.T) {
	const mockEntityOne string = "entityOne"
	const mockEntityTwo string = "entityTwo"
	const mockAttributeOne string = "mock_attribute_one"
	const mockAttributeTwo string = "mock_attribute_two"
	const stringValue string = "a value"
	const intValue int64 = int64(3)

	testDao := mapstore.NewEntityDaoMapImpl()
	testDao.SetAttribute(mockEntityOne,mockAttributeOne,stringValue)
	testDao.SetAttribute(mockEntityTwo,mockAttributeTwo,intValue)
	stringResult, err := testDao.GetAttribute(mockEntityOne,mockAttributeOne)
	if err != nil {
		t.Fatal(err)
	}
	intResult, err := testDao.GetAttribute(mockEntityTwo,mockAttributeTwo)
	if err != nil {
		t.Fatal(err)
	}
	nonResult, err := testDao.GetAttribute(mockEntityTwo,mockAttributeOne)
	if err != nil {
		t.Fatal(err)
	} 

	entitiesWithAttributeOne, err := testDao.GetEntitiesWithAttributeType(mockAttributeOne)
	if err != nil {
		t.Fatal(err)
	}

	if (stringResult != stringValue) {
		t.Fatalf("Attribute value should be %v but was %v", stringValue, stringResult)
	}
	if (intResult != intValue) {
		t.Fatalf("Attribute value should be %v but was %v", intValue, intResult)
	}
	if (nonResult != nil) {
		t.Fatalf("Attribute value should be nil but was %v", nonResult)
	}

	if len(entitiesWithAttributeOne) != 1 {
		t.Fatalf("expected 1 entity but got %v", len(entitiesWithAttributeOne))
	}
	if entitiesWithAttributeOne[0].GetID() != mockEntityOne {
		t.Fatalf("Exepected %v but got %v", mockEntityOne, entitiesWithAttributeOne[0].GetID())
	}

}

