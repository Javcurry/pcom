package util

import (
	"bytes"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// env val ...
var (
	mGoPath string
	mGoRoot string
	mGpLen  int
	mGrLen  int
)

// CallStack 生成可以用于log的调用栈信息（没有换行）
func CallStack(skip, depth int) string {
	if depth < 2 {
		depth = 2
	}
	if skip < 0 {
		skip = 0
	}

	var buf bytes.Buffer
	fpcs := make([]uintptr, depth)
	n := runtime.Callers(skip+2, fpcs) // +2的目的是跳出GenStack和Callers本身
	j := 0
	for i := n - 1; i >= 0; i-- {
		fun := runtime.FuncForPC(fpcs[i] - 1)
		if nil == fun {
			continue
		}
		fn := fun.Name()
		if strings.HasPrefix(fn, "runtime.") {
			continue
		}

		f, l := fun.FileLine(fpcs[i] - 1) // pc保存的是下一个地址，所以要回退
		if strings.HasPrefix(f, mGoPath) {
			f = f[mGpLen+1:]
		} else if strings.HasPrefix(f, mGoRoot) {
			f = f[mGrLen+1:]
		}

		if j > 0 {
			buf.WriteString(" --> ")
		}
		buf.WriteString(fn)
		buf.WriteString("(")
		buf.WriteString(f)
		buf.WriteString(":")
		buf.WriteString(strconv.Itoa(l))
		buf.WriteString(")")
		j++
	}

	return buf.String()
}

func init() {
	mGoPath = os.Getenv("GOPATH")
	mGoRoot = os.Getenv("GOROOT")
	mGpLen = len(mGoPath)
	mGrLen = len(mGoRoot)
}
