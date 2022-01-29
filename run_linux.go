package ffmpeg_go

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/u2takey/go-utils/rand"
)

const (
	cgroupConfigKey = "cgroupConfig"
	cpuRoot         = "/sys/fs/cgroup/cpu,cpuacct"
	cpuSetRoot      = "/sys/fs/cgroup/cpuset"
	procsFile       = "cgroup.procs"
	cpuSharesFile   = "cpu.shares"
	cfsPeriodUsFile = "cpu.cfs_period_us"
	cfsQuotaUsFile  = "cpu.cfs_quota_us"
	cpuSetCpusFile  = "cpuset.cpus"
	cpuSetMemsFile  = "cpuset.mems"
)

type cgroupConfig struct {
	cpuRequest float32
	cpuLimit   float32
	cpuset     string
	memset     string
}

func (s *Stream) setCGroupConfig(f func(config *cgroupConfig)) *Stream {
	a := s.Context.Value(cgroupConfigKey)
	if a == nil {
		a = &cgroupConfig{}
	}
	f(a.(*cgroupConfig))
	s.Context = context.WithValue(s.Context, cgroupConfigKey, a)
	return s
}

func (s *Stream) WithCpuCoreRequest(n float32) *Stream {
	return s.setCGroupConfig(func(config *cgroupConfig) {
		config.cpuRequest = n
	})
}

func (s *Stream) WithCpuCoreLimit(n float32) *Stream {
	return s.setCGroupConfig(func(config *cgroupConfig) {
		config.cpuLimit = n
	})
}

func (s *Stream) WithCpuSet(n string) *Stream {
	return s.setCGroupConfig(func(config *cgroupConfig) {
		config.cpuset = n
	})
}

func (s *Stream) WithMemSet(n string) *Stream {
	return s.setCGroupConfig(func(config *cgroupConfig) {
		config.memset = n
	})
}

func writeCGroupFile(rootPath, file string, value string) error {
	return ioutil.WriteFile(filepath.Join(rootPath, file), []byte(value), 0755)
}

func (s *Stream) RunWithResource(cpuRequest, cpuLimit float32) error {
	return s.WithCpuCoreRequest(cpuRequest).WithCpuCoreLimit(cpuLimit).RunLinux()
}

func (s *Stream) RunLinux() error {
	a := s.Context.Value(cgroupConfigKey).(*cgroupConfig)
	if a.cpuRequest > a.cpuLimit {
		return errors.New("cpuCoreLimit should greater or equal to cpuCoreRequest")
	}
	name := "ffmpeg_go_" + rand.String(6)
	rootCpuPath, rootCpuSetPath := filepath.Join(cpuRoot, name), filepath.Join(cpuSetRoot, name)
	err := os.MkdirAll(rootCpuPath, 0777)
	if err != nil {
		return err
	}
	err = os.MkdirAll(rootCpuSetPath, 0777)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(rootCpuPath); _ = os.Remove(rootCpuSetPath) }()

	share := int(1024 * a.cpuRequest)
	period := 100000
	quota := int(a.cpuLimit * 100000)

	if share > 0 {
		err = writeCGroupFile(rootCpuPath, cpuSharesFile, strconv.Itoa(share))
		if err != nil {
			return err
		}
	}
	err = writeCGroupFile(rootCpuPath, cfsPeriodUsFile, strconv.Itoa(period))
	if err != nil {
		return err
	}
	if quota > 0 {
		err = writeCGroupFile(rootCpuPath, cfsQuotaUsFile, strconv.Itoa(quota))
		if err != nil {
			return err
		}
	}
	if a.cpuset != "" && a.memset != "" {
		err = writeCGroupFile(rootCpuSetPath, cpuSetCpusFile, a.cpuset)
		if err != nil {
			return err
		}
		err = writeCGroupFile(rootCpuSetPath, cpuSetMemsFile, a.memset)
		if err != nil {
			return err
		}
	}

	cmd := s.Compile()
	err = cmd.Start()
	if err != nil {
		return err
	}
	if share > 0 || quota > 0 {
		err = writeCGroupFile(rootCpuPath, procsFile, strconv.Itoa(cmd.Process.Pid))
		if err != nil {
			return err
		}
	}
	if a.cpuset != "" && a.memset != "" {
		err = writeCGroupFile(rootCpuSetPath, procsFile, strconv.Itoa(cmd.Process.Pid))
		if err != nil {
			return err
		}
	}

	return cmd.Wait()
}
