package ffmpeg_go

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProbe(t *testing.T) {
	data, err := Probe(TestInputFile1, nil)
	assert.Nil(t, err)
	duration, err := probeOutputDuration(data)
	assert.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("%f", duration), "7.036000")
}

type probeFormat struct {
	Duration string `json:"duration"`
}

type probeData struct {
	Format probeFormat `json:"format"`
}

func probeOutputDuration(a string) (float64, error) {
	pd := probeData{}
	err := json.Unmarshal([]byte(a), &pd)
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(pd.Format.Duration, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}
