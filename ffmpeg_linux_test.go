package ffmpeg_go

import (
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileWithOptions(t *testing.T) {
	out := Input("dummy.mp4").Output("dummy2.mp4")
	cmd := out.Compile(SeparateProcessGroup())
	assert.Equal(t, 0, cmd.SysProcAttr.Pgid)
	assert.True(t, cmd.SysProcAttr.Setpgid)
}

func TestGlobalCommandOptions(t *testing.T) {
	out := Input("dummy.mp4").Output("dummy2.mp4")
	GlobalCommandOptions = append(GlobalCommandOptions, func(cmd *exec.Cmd) {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: 0}
	})
	defer func() {
		GlobalCommandOptions = GlobalCommandOptions[0 : len(GlobalCommandOptions)-1]
	}()
	cmd := out.Compile()
	assert.Equal(t, 0, cmd.SysProcAttr.Pgid)
	assert.True(t, cmd.SysProcAttr.Setpgid)
}
