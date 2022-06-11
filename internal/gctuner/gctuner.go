package gctuner

import (
	"context"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"github.com/spf13/afero"

	"github.com/zitryss/aye-and-nay/internal/log"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

const (
	cgroupMemTotalPathV1 = "/sys/fs/cgroup/memory/memory.limit_in_bytes"
	cgroupMemTotalPathV2 = "/sys/fs/cgroup/memory.max"
)

var (
	newGOGC       float64
	lastGOGC      float64
	memTotal      float64
	memLimitRatio = 0.7
	appFs         = afero.NewOsFs()
	memUsedFn     = memProcess
)

func Start(ctx context.Context, total int, ratio float64) error {
	if lastGOGC == 0.0 {
		gogc, ok := os.LookupEnv("GOGC")
		if !ok || gogc == "" {
			gogc = "100"
		}
		err := error(nil)
		lastGOGC, err = strconv.ParseFloat(gogc, 64)
		if err != nil {
			return errors.Wrap(err)
		}
	}
	if total > 0.0 {
		memTotal = float64(total)
	}
	if ratio > 0.0 && ratio <= 1.0 {
		memLimitRatio = ratio
	}
	err := checkMemTotal()
	if err != nil {
		return errors.Wrap(err)
	}
	err = checkCgroup()
	if err != nil {
		return errors.Wrap(err)
	}
	fin := &finalizer{}
	fin.ref = &finalizerRef{parent: fin}
	runtime.SetFinalizer(fin.ref, finalizerHandler(ctx))
	fin.ref = nil
	return nil
}

func checkMemTotal() error {
	if memTotal > 0.0 {
		return nil
	}
	memVirtual, err := mem.VirtualMemory()
	if err != nil {
		return errors.Wrap(err)
	}
	memTotal = float64(memVirtual.Total)
	return nil
}

func checkCgroup() error {
	var (
		f   io.ReadCloser
		err error
		e   error
		mt  float64
	)
	f, err = appFs.Open(cgroupMemTotalPathV1)
	if err != nil {
		e = errors.Wrap(err)
		goto second_file
	}
	defer f.Close()
	mt, err = readCgroupMemTotal(f)
	if err != nil {
		e = errors.Wrap(err)
		goto second_file
	}
	if mt > 0.0 && mt < memTotal {
		memTotal = mt
	}
second_file:
	f, err = appFs.Open(cgroupMemTotalPathV2)
	if err != nil {
		return nil
	}
	defer f.Close()
	mt, err = readCgroupMemTotal(f)
	if err != nil && e != nil {
		return errors.Wrapf(err, "%s", e)
	}
	if err != nil && e == nil {
		return nil
	}
	if mt > 0.0 && mt < memTotal {
		memTotal = mt
	}
	return nil
}

func readCgroupMemTotal(f io.Reader) (float64, error) {
	b, err := io.ReadAll(f)
	if err != nil {
		return 0.0, errors.Wrap(err)
	}
	s := strings.TrimSpace(string(b))
	if s == "" || s == "max" {
		return 0.0, nil
	}
	cgroupMemTotal, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, errors.Wrap(err)
	}
	if cgroupMemTotal <= 0.0 {
		return 0.0, nil
	}
	return cgroupMemTotal, nil
}

type finalizer struct {
	ref *finalizerRef
}

type finalizerRef struct {
	parent *finalizer
}

func finalizerHandler(ctx context.Context) func(fin *finalizerRef) {
	return func(fin *finalizerRef) {
		err := updateGOGC()
		if err != nil {
			log.Error(context.Background(), "err", err)
		}
		select {
		case <-ctx.Done():
			return
		default:
			runtime.SetFinalizer(fin, finalizerHandler(ctx))
		}
	}
}

func updateGOGC() error {
	memUsed, err := memUsedFn()
	if err != nil {
		return errors.Wrap(err)
	}
	if memTotal == 0 {
		return errors.New("division by zero")
	}
	memUsedRatio := memUsed / memTotal
	if memUsedRatio == 0 {
		return errors.New("division by zero")
	}
	newGOGC = (memLimitRatio - memUsedRatio) / memUsedRatio * 100.0
	if newGOGC < 0.0 {
		newGOGC = lastGOGC * memLimitRatio / memUsedRatio
	}
	lastGOGC = float64(debug.SetGCPercent(int(newGOGC)))
	log.Debug(context.Background(),
		"mem used", memUsed,
		"mem used ratio", memUsedRatio,
		"new GOGC", newGOGC,
	)
	return nil
}

func memProcess() (float64, error) {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return 0.0, errors.Wrap(err)
	}
	processMemory, err := p.MemoryInfo()
	if err != nil {
		return 0.0, errors.Wrap(err)
	}
	return float64(processMemory.RSS), nil
}
