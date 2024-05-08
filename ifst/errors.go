package ifst

import (
	"fmt"
	"reflect"
)

// invalidTransferInputError ...
type invalidTransferInputError struct {
	Type reflect.Type
}

func (e *invalidTransferInputError) Error() string {
	if e.Type == nil {
		return "ifst: Transfer(input: nil)"
	}
	return "ifst: Transfer(input: nil " + e.Type.String() + ")"
}

// invalidTransferOutputError ...
type invalidTransferOutputError struct {
	Type reflect.Type
}

func (e *invalidTransferOutputError) Error() string {
	if e.Type == nil {
		return "ifst: Transfer(output: nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "ifst: Transfer(output: non-pointer " + e.Type.String() + ")"
	}
	return "ifst: Transfer(output: nil " + e.Type.String() + ")"
}

// transferTypeError ...
type transferTypeError struct {
	Input  reflect.Type
	Output reflect.Type
}

func (e *transferTypeError) Error() string {
	return fmt.Sprintf("ifst: cannot transfer %v into %v", e.Input.String(), e.Output.String())
}

// transferInvalidMapKeyTypeError
type transferInvalidMapKeyTypeError struct {
	Type reflect.Type
}

func (e *transferInvalidMapKeyTypeError) Error() string {
	return fmt.Sprintf("ifst: invalid key type of map: %v", e.Type.String())
}

// structTagInvalidError ...
type structTagInvalidError struct {
}

func (e *structTagInvalidError) Error() string {
	return "ifst: invalid tag of struct"
}

// marshalerError ...
type marshalerError struct {
	Type reflect.Type
	Err  error
}

func (e *marshalerError) Error() string {
	return "ifst: error calling MarshalIFST for type " + e.Type.String() + ": " + e.Err.Error()
}

// unmarshalerError ...
type unmarshalerError struct {
	Type reflect.Type
	Err  error
}

func (e *unmarshalerError) Error() string {
	return "ifst: error calling UnmarshalIFST for type " + e.Type.String() + ": " + e.Err.Error()
}

// weakTypeTransferError
type weakTypeTransferError struct {
	Input  reflect.Type
	Output reflect.Type
}

func (e *weakTypeTransferError) Error() string {
	return fmt.Sprintf("ifst: weak type transfer %v into %v disabled, enable WeakTypeTransfer before transfer if you need", e.Input.String(), e.Output.String())
}
