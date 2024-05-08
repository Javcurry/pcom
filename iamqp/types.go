package iamqp

import (
	"errors"
	"fmt"
)

// Log ...
var Log = func(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// LogFunc ...
type LogFunc func(format string, args ...interface{})

// SetLogger set logger for this library
func SetLogger(log LogFunc) {
	Log = log
}

// Err definitions
var (
	ErrShutDown               = errors.New("shut down")
	errDeliveryNotInitialized = errors.New("delivery not initialized")
)
