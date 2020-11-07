package store_test

import (
	"projects/parser/internal/model"
	"projects/parser/internal/store"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNews_Create(t *testing.T) {
	s, tearDown := store.TestStore(t, "", "")
	defer tearDown("news")

	new, err := s.News().Create(&model.News{
		Url: "testUrl",
	})
	assert.NoError(t, err)
	assert.NotNil(t, new)
	new, err = s.News().Create(&model.News{
		Url: "testUrl",
	})
	assert.NoError(t, err)
	assert.NotNil(t, new)
}

func TestNews_DeleteByUrl(t *testing.T) {
	s, tearDown := store.TestStore(t, "", "")
	defer tearDown("news")

	new, err := s.News().Create(&model.News{
		Url: "testUrl",
	})
	assert.NoError(t, err)
	assert.NotNil(t, new)
	err = s.News().DeleteByUrl("testUrl")
	assert.NoError(t, err)
}

func TestNews_GetAll(t *testing.T) {
	s, tearDown := store.TestStore(t, "", "")
	defer tearDown("news")

	new, err := s.News().Create(&model.News{
		Url: "testUrl1",
	})
	assert.NoError(t, err)
	assert.NotNil(t, new)
	new, err = s.News().Create(&model.News{
		Url: "testUrl2",
	})
	assert.NoError(t, err)
	assert.NotNil(t, new)
	new, err = s.News().Create(&model.News{
		Url: "testUrl3",
	})
	assert.NoError(t, err)
	assert.NotNil(t, new)
	news, err := s.News().GetAll()
	assert.NoError(t, err)
	result := ""
	for _, v := range news {
		result = result + v.Url + ","
	}
	assert.Equal(t, "testUrl1,testUrl2,testUrl3,", result)
}
