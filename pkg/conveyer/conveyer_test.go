package conveyer

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConveyer_google(t *testing.T) {
	wg := sync.WaitGroup{}
	test, err := New("ifconfig.co", nil)
	assert.Nil(t, err)
	assert.NotNil(t, test)
	go test.Start(context.Background(), &wg)
	test.GetInput() <- []byte("https://ifconfig.co/country")
	result := string(<-test.GetOutput())
	assert.Equal(t, result, "Russia\n")
	test.Close()
}
