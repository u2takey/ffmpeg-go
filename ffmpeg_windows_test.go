package ffmpeg_go

import (
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileWithOptions(t *testing.T) {
	out := Input("dummy.mp4").Output("dummy2.mp4")
	cmd := out.Compile(func(s *Stream, cmd *exec.Cmd) {
		if cmd.SysProcAttr == nil {
			cmd.SysProcAttr = &syscall.SysProcAttr{}
		}
		cmd.SysProcAttr.HideWindow = true
	})
	assert.Equal(t, true, cmd.SysProcAttr.HideWindow)
}
