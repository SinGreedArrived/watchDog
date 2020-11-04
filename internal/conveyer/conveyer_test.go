package conveyer

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConveyer_google(t *testing.T) {
	wg := sync.WaitGroup{}
	test := New("google", nil)
	assert.Nil(t, test)
	assert.NotNil(t, test.Start(&wg))
	test.GetInputChan() <- "google.com"
	assert.Equal(t, <-test.GetOutputChan(), []byte("google"))
	test.Close()
}
