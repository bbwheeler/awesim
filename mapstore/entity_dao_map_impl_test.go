package dao_test

import "testing"
import mapstore "github.com/bbwheeler/awesim/mapstore"
import "reflect"
import "fmt"
import "sort"
import "github.com/google/uuid"

func TestRemoveEntity(t *testing.T) {
	const testAttribute string = "testAttribute"
	const testAttributeValue string = "test"
	testDao := mapstore.NewEntityDaoMapImpl()

	entityID := uuid.New().String()
	
	err := testDao.SetAttribute(entityID, testAttribute, testAttributeValue)
	if err != nil {
		t.Fatal(err)
	}

	att, err := testDao.GetAttribute(entityID, testAttribute)
	if err != nil {
		t.Fatal(err)
	}
	if att != testAttributeValue {
		t.Fatalf("expected value to be %s but was %v", testAttributeValue, att)
	}

	err = testDao.RemoveEntity(entityID)
	if err != nil {
		t.Fatal(err)
	}

	att, err = testDao.GetAttribute(entityID, testAttribute)
	if err == nil {
		t.Fatal(err)
	}
}

func TestSetHasAndGetAttribute(t *testing.T) {

	tests := []struct {
		name string
		attributeValue interface{}
		doNotSetAttributeValue bool
		expectSetError bool
		expectGetError bool
	} {
		{
			name: "string",
			attributeValue: "a string",
		},
		{
			name: "int64",
			attributeValue: int64(4),
		},
		{
			name: "float64",
			attributeValue: float64(1.1),
		},
		{
			name: "boolean",
			attributeValue: true,
		},
		{
			name: "uint64", // unsupported
			attributeValue: uint64(5),
			expectSetError: true,
		},
		{
			name: "nil", // unsupported
			expectSetError: true,
		},
		{
			name: "no attribute",
			doNotSetAttributeValue: true,
			expectGetError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const mockEntityOne string = "entityOne"
			const mockAttributeOne string = "mock_attribute_one"
		
			testDao := mapstore.NewEntityDaoMapImpl()
			var err error
			if !test.doNotSetAttributeValue {
				err = testDao.SetAttribute(mockEntityOne,mockAttributeOne,test.attributeValue)
			}			
			if err != nil {
				if test.expectSetError {
					return
				}
				t.Fatal(err)
			}
			if test.expectSetError {
				t.Fatalf("Expected error when setting attribute")
			}

			hasAttribute, err := testDao.HasAttribute(mockEntityOne, mockAttributeOne)
			if err != nil {
				t.Fatal("Error from HasAttribute")
			}
			if hasAttribute == test.doNotSetAttributeValue {
				t.Fatalf("HasAttribute was %v but expected %v", hasAttribute, !hasAttribute)
			}

			val, err := testDao.GetAttribute(mockEntityOne,mockAttributeOne)
			if err != nil {
				if test.expectGetError {
					return
				}
				t.Fatal(err)
			}
			if val != test.attributeValue {
				t.Fatalf("Expected value to be %v but was %v", test.attributeValue, val)
			}
		})
	}

}

func TestGetEntitiesWithAttributeType(t *testing.T) {

	tests := []struct {
		name string
		numberOfEntitiesWithAttribute int
		numberOfEntitiesWithoutAttribute int
	} {
		{
			name: "one attribute",
			numberOfEntitiesWithAttribute: 1,
			numberOfEntitiesWithoutAttribute: 0,
		},
		{
			name: "no attributes",
			numberOfEntitiesWithAttribute: 0,
			numberOfEntitiesWithoutAttribute: 0,
		},
		{
			name: "only other attributes",
			numberOfEntitiesWithAttribute: 0,
			numberOfEntitiesWithoutAttribute: 1,
		},
		{
			name: "one and one",
			numberOfEntitiesWithAttribute: 1,
			numberOfEntitiesWithoutAttribute: 1,
		},
		{
			name: "two attribute",
			numberOfEntitiesWithAttribute: 2,
			numberOfEntitiesWithoutAttribute: 0,
		},
		{
			name: "two and one",
			numberOfEntitiesWithAttribute: 2,
			numberOfEntitiesWithoutAttribute: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const entityPrefix string = "test"
		
			testDao := mapstore.NewEntityDaoMapImpl()
			var entitiesWithAttribute []string

			testAttribute := "mock attribute"			
			for i := 0; i < test.numberOfEntitiesWithAttribute; i++ {
				entityID := fmt.Sprintf("%s%d",entityPrefix,i)
				entitiesWithAttribute = append(entitiesWithAttribute, entityID) 
				testDao.SetAttribute(entityID,testAttribute,int64(i))
				if i%2 != 0 {
					testDao.SetAttribute(entityID,"dummyAttribute",int64(i+1))
				}
			}
			for i := test.numberOfEntitiesWithAttribute; i < test.numberOfEntitiesWithAttribute + test.numberOfEntitiesWithoutAttribute; i++ {
				testDao.SetAttribute(fmt.Sprintf("%s%d",entityPrefix,i),"dummyAttribute",int64(i+1))
			}

			resultIDs, err := testDao.GetEntitiesWithAttributeType(testAttribute)
			if err != nil {
				t.Fatal(err)
			}
			if len(resultIDs) != test.numberOfEntitiesWithAttribute {
				t.Fatalf("Expected entities list with %d values but got %d values", test.numberOfEntitiesWithAttribute, len(resultIDs))
			}

			if len(resultIDs) != test.numberOfEntitiesWithAttribute {
				t.Fatalf("Expected %d results but got %d", test.numberOfEntitiesWithAttribute, len(resultIDs))
			}

			sort.Strings(resultIDs)
			sort.Strings(entitiesWithAttribute)

			if !reflect.DeepEqual(resultIDs, entitiesWithAttribute) {
				t.Fatalf("expected %v but got %v", entitiesWithAttribute, resultIDs)
			}		
		})
	}
}

func TestGetEntitiesWithAttribute(t *testing.T) {

	tests := []struct {
		name string
		numberOfEntitiesWithAttributeAndValue int
		numberOfEntitiesWithAttributeButNotValue int
		numberOfEntitiesWithoutAttribute int
	} {
		{
			name: "one attribute",
			numberOfEntitiesWithAttributeAndValue: 1,
			numberOfEntitiesWithAttributeButNotValue: 0,			
			numberOfEntitiesWithoutAttribute: 0,
		},
		{
			name: "no attributes",
			numberOfEntitiesWithAttributeAndValue: 0,
			numberOfEntitiesWithAttributeButNotValue: 0,
			numberOfEntitiesWithoutAttribute: 0,
		},
		{
			name: "only other attributes",
			numberOfEntitiesWithAttributeAndValue: 0,
			numberOfEntitiesWithAttributeButNotValue: 0,
			numberOfEntitiesWithoutAttribute: 1,
		},
		{
			name: "one attribute one not",
			numberOfEntitiesWithAttributeAndValue: 1,
			numberOfEntitiesWithAttributeButNotValue: 0,
			numberOfEntitiesWithoutAttribute: 1,
		},
		{
			name: "two attributes",
			numberOfEntitiesWithAttributeAndValue: 2,
			numberOfEntitiesWithAttributeButNotValue: 0,
			numberOfEntitiesWithoutAttribute: 0,
		},
		{
			name: "two attributes one not",
			numberOfEntitiesWithAttributeAndValue: 2,
			numberOfEntitiesWithAttributeButNotValue: 0,
			numberOfEntitiesWithoutAttribute: 1,
		},
		{
			name: "one attribute with different value",
			numberOfEntitiesWithAttributeAndValue: 0,
			numberOfEntitiesWithAttributeButNotValue: 1,
			numberOfEntitiesWithoutAttribute: 0,
		},
		{
			name: "one attribute with correct value and one with different value",
			numberOfEntitiesWithAttributeAndValue: 1,
			numberOfEntitiesWithAttributeButNotValue: 1,
			numberOfEntitiesWithoutAttribute: 0,
		},
		{
			name: "one of each",
			numberOfEntitiesWithAttributeAndValue: 1,
			numberOfEntitiesWithAttributeButNotValue: 1,
			numberOfEntitiesWithoutAttribute: 1,
		},
		{
			name: "three of each",
			numberOfEntitiesWithAttributeAndValue: 3,
			numberOfEntitiesWithAttributeButNotValue: 3,
			numberOfEntitiesWithoutAttribute: 3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const entityPrefix string = "test"
		
			testDao := mapstore.NewEntityDaoMapImpl()
			var entitiesWithAttribute []string

			testAttribute := "mock attribute"
			testAttributeValue := "mock value"			
			incorrectAttributeValue := "wrong value"			
			for i := 0; i < test.numberOfEntitiesWithAttributeAndValue; i++ {
				entityID := fmt.Sprintf("%s%d",entityPrefix,i)
				entitiesWithAttribute = append(entitiesWithAttribute, entityID) 
				testDao.SetAttribute(entityID,testAttribute,testAttributeValue)
				if i%2 != 0 {
					testDao.SetAttribute(entityID,"dummyAttribute",int64(i+1))
				}
			}
			for i := test.numberOfEntitiesWithAttributeAndValue; i < test.numberOfEntitiesWithAttributeAndValue + test.numberOfEntitiesWithAttributeButNotValue; i++ {
				testDao.SetAttribute(fmt.Sprintf("%s%d",entityPrefix,i),testAttribute,incorrectAttributeValue)
			}
			for i := test.numberOfEntitiesWithAttributeAndValue + test.numberOfEntitiesWithAttributeButNotValue; i < test.numberOfEntitiesWithAttributeAndValue + test.numberOfEntitiesWithAttributeButNotValue + test.numberOfEntitiesWithoutAttribute; i++ {
				testDao.SetAttribute(fmt.Sprintf("%s%d",entityPrefix,i),"dummyAttribute",int64(i+1))
			}

			resultIDs, err := testDao.GetEntitiesWithAttribute(testAttribute, testAttributeValue)
			if err != nil {
				t.Fatal(err)
			}
			if len(resultIDs) != test.numberOfEntitiesWithAttributeAndValue {
				t.Fatalf("Expected entities list with %d values but got %d values", test.numberOfEntitiesWithAttributeAndValue, len(resultIDs))
			}

			if len(resultIDs) != test.numberOfEntitiesWithAttributeAndValue {
				t.Fatalf("Expected %d results but got %d", test.numberOfEntitiesWithAttributeAndValue, len(resultIDs))
			}

			sort.Strings(resultIDs)
			sort.Strings(entitiesWithAttribute)

			if !reflect.DeepEqual(resultIDs, entitiesWithAttribute) {
				t.Fatalf("expected %v but got %v", entitiesWithAttribute, resultIDs)
			}		
		})
	}
}



func TestEndToEnd(t *testing.T) {
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
	_, err = testDao.GetAttribute(mockEntityTwo,mockAttributeOne)
	if err == nil {
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

	if len(entitiesWithAttributeOne) != 1 {
		t.Fatalf("expected 1 entity but got %v", len(entitiesWithAttributeOne))
	}
	if entitiesWithAttributeOne[0] != mockEntityOne {
		t.Fatalf("Exepected %v but got %v", mockEntityOne, entitiesWithAttributeOne[0])
	}

}

