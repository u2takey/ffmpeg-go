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

func TestGlobalCommandOptions(t *testing.T) {
	out := Input("dummy.mp4").Output("dummy2.mp4")
	GlobalCommandOptions = append(GlobalCommandOptions, func(cmd *exec.Cmd) {
		if cmd.SysProcAttr == nil {
			cmd.SysProcAttr = &syscall.SysProcAttr{}
		}
		cmd.SysProcAttr.HideWindow = true
	})
	defer func() {
		GlobalCommandOptions = GlobalCommandOptions[0 : len(GlobalCommandOptions)-1]
	}()
	cmd := out.Compile()
	assert.Equal(t, true, cmd.SysProcAttr.HideWindow)
}
