package ffmpeg_go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileWithOptions(t *testing.T) {
	out := Input("dummy.mp4").Output("dummy2.mp4")
	cmd := out.Compile(SeparateProcessGroup())
	assert.Equal(t, cmd.SysProcAttr.Pgid, 0)
	assert.True(t, cmd.SysProcAttr.Setpgid)
}
