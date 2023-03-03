package policyTester

import (
	"bufio"
	"io"
	"reflect"
	"regexp"
	"runtime/pprof"
	"strings"
	"sync"
	"time"
)

// TestDeps is an implementation of the testing.testDeps interface,
// suitable for passing to testing.MainStart.
type TestDeps struct{}

var matchPat string
var matchRe *regexp.Regexp

func (TestDeps) MatchString(pat, str string) (result bool, err error) {
	if matchRe == nil || matchPat != pat {
		matchPat = pat
		matchRe, err = regexp.Compile(matchPat)
		if err != nil {
			return
		}
	}
	return matchRe.MatchString(str), nil
}

func (TestDeps) StartCPUProfile(w io.Writer) error {
	return pprof.StartCPUProfile(w)
}

func (TestDeps) StopCPUProfile() {
	pprof.StopCPUProfile()
}

func (TestDeps) WriteProfileTo(name string, w io.Writer, debug int) error {
	return pprof.Lookup(name).WriteTo(w, debug)
}

// ImportPath is the import path of the testing binary, set by the generated main function.
var ImportPath string

func (TestDeps) ImportPath() string {
	return ImportPath
}

// testLog implements testlog.Interface, logging actions by package os.
type testLog struct {
	mu sync.Mutex
	w  *bufio.Writer
}

func (l *testLog) Getenv(key string) {
	l.add("getenv", key)
}

func (l *testLog) Open(name string) {
	l.add("open", name)
}

func (l *testLog) Stat(name string) {
	l.add("stat", name)
}

func (l *testLog) Chdir(name string) {
	l.add("chdir", name)
}

// add adds the (op, name) pair to the test log.
func (l *testLog) add(op, name string) {
	if strings.Contains(name, "\n") || name == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.w == nil {
		return
	}
	l.w.WriteString(op)
	l.w.WriteByte(' ')
	l.w.WriteString(name)
	l.w.WriteByte('\n')
}

var logs testLog

func (TestDeps) StartTestLog(w io.Writer) {}

func (TestDeps) StopTestLog() error {
	logs.mu.Lock()
	defer logs.mu.Unlock()
	err := logs.w.Flush()
	logs.w = nil
	return err
}

func (TestDeps) SetPanicOnExit0(bool) {}

func (TestDeps) CoordinateFuzzing(time.Duration, int64, time.Duration, int64, int, []struct {
	Parent     string
	Path       string
	Data       []byte
	Values     []any
	Generation int
	IsSeed     bool
}, []reflect.Type, string, string) (err error) {
	return err
}

func (TestDeps) RunFuzzWorker(func(struct {
	Parent     string
	Path       string
	Data       []byte
	Values     []any
	Generation int
	IsSeed     bool
}) error) error {
	return nil
}

func (TestDeps) ReadCorpus(string, []reflect.Type) ([]struct {
	Parent     string
	Path       string
	Data       []byte
	Values     []any
	Generation int
	IsSeed     bool
}, error) {
	return nil, nil
}

func (TestDeps) CheckCorpus(vals []any, types []reflect.Type) error {
	return nil
}

func (TestDeps) ResetCoverage() {
}

func (TestDeps) SnapshotCoverage() {
}
