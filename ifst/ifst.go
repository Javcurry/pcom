package ifst

import (
	"fmt"
	"reflect"
	"runtime"
)

// Transformer for transform data
type Transformer struct {
	TagName          string
	WeakTypeTransfer bool
}

// NewTransformer constructor
func NewTransformer() *Transformer {
	return &Transformer{
		TagName: "ifst",
	}
}

// EnableWeakTypeTransfer ...
func (tf *Transformer) EnableWeakTypeTransfer() {
	tf.WeakTypeTransfer = true
}

// DisableWeakTypeTransfer ...
func (tf *Transformer) DisableWeakTypeTransfer() {
	tf.WeakTypeTransfer = false
}

// Transfer data from input into output
func (tf *Transformer) Transfer(input interface{}, output interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}
			err = r.(error)
		}
	}()

	return tf.reflectTransfer(input, output)
}

func (tf *Transformer) reflectTransfer(input interface{}, output interface{}) (err error) {
	iv := reflect.ValueOf(input)
	ov := reflect.ValueOf(output)

	if !iv.IsValid() || !ov.IsValid() {
		return fmt.Errorf("input or output is invalid")
	}

	if iv.Kind() == reflect.Ptr && iv.IsNil() {
		return &invalidTransferInputError{Type: iv.Type()}
	}
	if ov.Kind() != reflect.Ptr || ov.IsNil() {
		return &invalidTransferOutputError{Type: ov.Type()}
	}

	return reflectValueTransfer(iv, ov, &transferOptions{
		Transformer: tf,
	})
}

// Transfer from input into output by default settings
func Transfer(input interface{}, output interface{}) error {
	tr := NewTransformer()
	return tr.Transfer(input, output)
}

// TransferWithTagName from input into output with specified tag name
func TransferWithTagName(input interface{}, output interface{}, tagName string) error {
	tr := NewTransformer()
	tr.TagName = tagName
	return tr.Transfer(input, output)
}
