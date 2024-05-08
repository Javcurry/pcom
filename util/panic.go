package util

import (
	"fmt"

	"git.ihago.top/ihago/ylog"
)

// Panic ...
func Panic() bool {
	err := recover()
	if err == nil {
		return false
	}
	ylog.Error(fmt.Sprintf("panic: err=%v. stack=[%v]", err, CallStack(1, 10))) // 1的目的是跳过Panic本身
	return true
}
