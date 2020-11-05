package store_test

import (
	"projects/parser/internal/model"
	"projects/parser/internal/store"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTarget_Create(t *testing.T) {
	s, tearDown := store.TestStore(t, "", "")
	defer tearDown("targets")

	target, err := s.Target().Create(&model.Target{
		Url:  "testUrl",
		Hash: "HASHING",
	})
	assert.NoError(t, err)
	assert.NotNil(t, target)
	target, err = s.Target().Create(&model.Target{
		Url:  "testUrl",
		Hash: "HASH2",
	})
	assert.NoError(t, err)
	assert.NotNil(t, target)
	target, err = s.Target().Create(&model.Target{
		Url:  "testUrl",
		Hash: "HASH2",
	})
	assert.NoError(t, err)
	assert.NotNil(t, target)
}

func TestTarget_FindByUrl(t *testing.T) {
	s, tearDown := store.TestStore(t, "", "")
	defer tearDown("targets")

	_, err := s.Target().FindByUrl("unlucky")
	assert.Error(t, err)

	target, err := s.Target().Create(&model.Target{
		Url:  "testUrl",
		Hash: "HASHING",
	})
	assert.NoError(t, err)
	assert.NotNil(t, target)

	target, err = s.Target().FindByUrl(target.Url)
	assert.NoError(t, err)
	assert.NotNil(t, target)
}

func TestTarget_Update(t *testing.T) {
	s, tearDown := store.TestStore(t, "", "")
	defer tearDown("targets")

	target, err := s.Target().Create(&model.Target{
		Url:  "testUrl",
		Hash: "HASHING",
	})
	assert.NoError(t, err)
	assert.NotNil(t, target)

	err = s.Target().Update(&model.Target{
		Url:  "testUrl",
		Hash: "HASH2",
	})
	assert.NoError(t, err)

	target, err = s.Target().FindByUrl("testUrl")
	assert.NoError(t, err)
	assert.NotNil(t, target)
	assert.Equal(t, target.Hash, "HASH2")
}
