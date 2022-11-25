package ffmpeg_go

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"time"
)

// ProbeReader** functions are the same as Probe** but accepting io.Reader instead of fileName

// ProbeReader runs ffprobe passing given reader via stdin and return a JSON representation of the output.
func ProbeReader(r io.Reader, kwargs ...KwArgs) (string, error) {
	return ProbeReaderWithTimeout(r, 0, MergeKwArgs(kwargs))
}

func ProbeReaderWithTimeout(r io.Reader, timeOut time.Duration, kwargs KwArgs) (string, error) {
	args := KwArgs{
		"show_format":  "",
		"show_streams": "",
		"of":           "json",
	}

	return ProbeReaderWithTimeoutExec(r, timeOut, MergeKwArgs([]KwArgs{args, kwargs}))
}

func ProbeReaderWithTimeoutExec(r io.Reader, timeOut time.Duration, kwargs KwArgs) (string, error) {
	args := ConvertKwargsToCmdLineArgs(kwargs)
	args = append(args, "-")
	ctx := context.Background()
	if timeOut > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), timeOut)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	cmd.Stdin = r
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return string(buf.Bytes()), nil
}
