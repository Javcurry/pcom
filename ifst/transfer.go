package ifst

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"gopkg.in/mgo.v2/bson"
)

// Marshaler ifst marshaler
type Marshaler interface {
	MarshalIFST() (interface{}, error)
}

// Unmarshaler ifst unmarshaler
type Unmarshaler interface {
	UnmarshalIFST(input interface{}) error
}

var (
	marshalerType     = reflect.TypeOf(new(Marshaler)).Elem()
	unmarshalerType   = reflect.TypeOf(new(Unmarshaler)).Elem()
	textMarshalerType = reflect.TypeOf(new(encoding.TextMarshaler)).Elem()

	jsonMarshalerType   = reflect.TypeOf(new(json.Marshaler)).Elem()
	jsonUnmarshalerType = reflect.TypeOf(new(json.Unmarshaler)).Elem()

	bsonSetterType = reflect.TypeOf(new(bson.Setter)).Elem()
	bsonGetterType = reflect.TypeOf(new(bson.Getter)).Elem()
)

type transferFunc func(reflect.Value, reflect.Value, *transferOptions) error

var literalKindUpgradeTable = map[reflect.Kind]reflect.Kind{
	reflect.Bool:   reflect.Bool,
	reflect.String: reflect.String,

	reflect.Int:   reflect.Int64,
	reflect.Int8:  reflect.Int64,
	reflect.Int16: reflect.Int64,
	reflect.Int32: reflect.Int64,
	reflect.Int64: reflect.Int64,

	reflect.Uint:    reflect.Uint64,
	reflect.Uint8:   reflect.Uint64,
	reflect.Uint16:  reflect.Uint64,
	reflect.Uint32:  reflect.Uint64,
	reflect.Uint64:  reflect.Uint64,
	reflect.Uintptr: reflect.Uint64,

	reflect.Float32: reflect.Float64,
	reflect.Float64: reflect.Float64,
}

var typeTransferTable = map[reflect.Kind]map[reflect.Kind]transferFunc{}

type transferOptions struct {
	Transformer *Transformer
}

func reflectValueTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !input.IsValid() || !output.IsValid() {
		return nil
	}

	inputType := input.Type()
	outputType := output.Type()

	if inputType.Implements(marshalerType) {
		return marshalerTransfer(input, output, opt)
	}
	if outputType.Implements(unmarshalerType) {
		return unmarshalerTransfer(input, output, opt)
	}

	if opt.Transformer.TagName == "json" { // Json Marshaler and Unmarshaler
		if inputType.Implements(jsonMarshalerType) {
			return jsonMarshalerTransfer(input, output, opt)
		}
		if outputType.Implements(jsonUnmarshalerType) {
			return jsonUnmarshalerTransfer(input, output, opt)
		} else if outputType.Kind() != reflect.Ptr && output.CanAddr() && reflect.PtrTo(outputType).Implements(jsonUnmarshalerType) {
			return jsonUnmarshalerTransfer(input, output.Addr(), opt)
		}
	} else if opt.Transformer.TagName == "bson" { // Bson Getter and Setter
		if inputType.Implements(bsonGetterType) {
			return bsonMarshalerTransfer(input, output, opt)
		}
		if outputType.Implements(bsonSetterType) {
			return bsonUnmarshalerTransfer(input, output, opt)
		} else if outputType.Kind() != reflect.Ptr && output.CanAddr() && reflect.PtrTo(outputType).Implements(jsonUnmarshalerType) {
			return bsonUnmarshalerTransfer(input, output.Addr(), opt)
		}
	}

	output = reflect.Indirect(output)
	fn := getTransferFunc(input, output)
	if fn == nil {
		return &transferTypeError{Input: inputType, Output: outputType}
	}
	return fn(input, output, opt)
}

func marshalerTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	vi := input.Interface().(Marshaler)
	marshalData, err := vi.MarshalIFST()
	if err != nil {
		return &marshalerError{Type: input.Type(), Err: err}
	}
	return reflectValueTransfer(reflect.ValueOf(marshalData), output, opt)
}

func unmarshalerTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	vo := output.Interface().(Unmarshaler)
	err := vo.UnmarshalIFST(input.Interface())
	if err != nil {
		return &unmarshalerError{Type: output.Type(), Err: err}
	}

	return nil
}

func jsonMarshalerTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	vi := input.Interface().(json.Marshaler)
	marshalData, err := vi.MarshalJSON()
	if err != nil {
		return &marshalerError{Type: input.Type(), Err: err}
	}

	var marshalInput interface{}
	err = json.Unmarshal(marshalData, &marshalInput)
	if err != nil {
		return err
	}

	return reflectValueTransfer(reflect.ValueOf(marshalInput), output, opt)
}

func jsonUnmarshalerTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	data, err := json.Marshal(input.Interface())
	if err != nil {
		return err
	}

	vo := output.Interface().(json.Unmarshaler)
	err = vo.UnmarshalJSON(data)
	if err != nil {
		return &unmarshalerError{Type: output.Type(), Err: err}
	}

	return nil
}

func bsonMarshalerTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	vi := input.Interface().(bson.Getter)
	marshalData, err := vi.GetBSON()
	if err != nil {
		return &marshalerError{Type: input.Type(), Err: err}
	}

	return reflectValueTransfer(reflect.ValueOf(marshalData), output, opt)
}

func bsonUnmarshalerTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	data, err := bson.Marshal(input.Interface())
	if err != nil {
		return err
	}

	var raw bson.Raw
	err = bson.Unmarshal(data, &raw)
	if err != nil {
		return &unmarshalerError{Type: output.Type(), Err: err}
	}

	vo := output.Interface().(bson.Setter)
	err = vo.SetBSON(raw)
	if err != nil {
		return &unmarshalerError{Type: output.Type(), Err: err}
	}

	return nil
}

func getTransferFunc(input reflect.Value, output reflect.Value) transferFunc {
	inputKind, exists := literalKindUpgradeTable[input.Kind()]
	if !exists {
		inputKind = input.Kind()
	}

	outputKind, exists := literalKindUpgradeTable[output.Kind()]
	if !exists {
		outputKind = output.Kind()
	}

	outputTransfer, exists := typeTransferTable[inputKind]
	if !exists {
		return nil
	}

	fn, exists := outputTransfer[outputKind]
	if !exists {
		fn, exists = outputTransfer[reflect.Invalid] // reflect.Invalid represent all type
		if !exists {
			return nil
		}
		return fn
	}

	return fn
}

func literalToInterfaceLiteralTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	output.Set(input)
	return nil
}

func intToIntTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	output.SetInt(input.Int())
	return nil
}

func intToUintTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	output.SetUint(uint64(input.Int()))
	return nil
}

func intToFloatTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	output.SetFloat(float64(input.Int()))
	return nil
}

func intToStringTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	output.SetString(strconv.FormatInt(input.Int(), 10))
	return nil
}

func uintToIntTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	output.SetInt(int64(input.Uint()))
	return nil
}

func uintToUintTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	output.SetUint(input.Uint())
	return nil
}

func uintToFloatTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	output.SetFloat(float64(input.Uint()))
	return nil
}

func uintToStringTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	output.SetString(strconv.FormatUint(input.Uint(), 10))
	return nil
}

func floatToFloatTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	output.SetFloat(input.Float())
	return nil
}

func floatToIntTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	output.SetInt(int64(input.Float()))
	return nil
}

func floatToUintTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	output.SetUint(uint64(input.Float()))
	return nil
}

func floatToStringTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	str := strconv.FormatFloat(input.Float(), 'f', -1, 64)

	output.SetString(str)
	return nil
}

func stringToStringTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	output.SetString(input.String())
	return nil
}

func stringToIntTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	number, err := strconv.ParseInt(input.String(), 10, 64)
	if err != nil {
		return err
	}
	output.SetInt(number)
	return nil
}

func stringToUintTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	number, err := strconv.ParseUint(input.String(), 10, 64)
	if err != nil {
		return err
	}
	output.SetUint(number)
	return nil
}

func stringToFloatTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if !opt.Transformer.WeakTypeTransfer {
		return &weakTypeTransferError{
			Input:  input.Type(),
			Output: output.Type(),
		}
	}

	float, err := strconv.ParseFloat(input.String(), 64)
	if err != nil {
		return err
	}
	output.SetFloat(float)
	return nil
}

var sliceToInterfaceOutputType = reflect.TypeOf([]interface{}{})

func sliceToInterfaceTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	in := input.Len()
	if in == 0 {
		return nil
	}

	o := reflect.New(sliceToInterfaceOutputType).Elem()
	err := sliceToSliceTransfer(input, o, opt)
	if err != nil {
		return err
	}

	output.Set(o)
	return nil
}

func sliceToSliceTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	in := input.Len()
	if in == 0 {
		return nil
	}

	if output.Cap() < in {
		newSlice := reflect.MakeSlice(output.Type(), 0, in)
		output.Set(newSlice)
	}
	output.SetLen(in)

	for i := 0; i < in; i++ {
		err := reflectValueTransfer(input.Index(i), output.Index(i), opt)
		if err != nil {
			return err
		}
	}

	return nil
}

func arrayToArrayTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	in := input.Len()
	if in == 0 {
		return nil
	}

	on := output.Len()
	for i := 0; i < in && i < on; i++ {
		err := reflectValueTransfer(input.Index(i), output.Index(i), opt)
		if err != nil {
			return err
		}
	}
	for i := in; i < on; i++ {
		output.Index(i).Set(reflect.Zero(output.Type().Elem()))
	}

	return nil
}

func arrayToStructTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	in := input.Len()
	if in == 0 {
		return nil
	}

	for i := 0; i < in; i++ {
		err := reflectValueTransfer(input.Index(i), output.Field(i), opt)
		if err != nil {
			return err
		}
	}

	return nil
}

func interfaceTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	if input.IsNil() {
		return nil
	}

	return reflectValueTransfer(input.Elem(), output, opt)
}

func ptrTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	return reflectValueTransfer(input.Elem(), output, opt)
}

func mapKeyValidate(t reflect.Type) error {
	switch t.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		if !t.Implements(textMarshalerType) {
			return &transferInvalidMapKeyTypeError{Type: t}
		}
	}

	return nil
}

func mapKeyMarshalText(v reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	}

	vType := v.Type()
	if vType.Implements(textMarshalerType) {
		vi := v.Interface().(encoding.TextMarshaler)
		key, err := vi.MarshalText()
		if err != nil {
			return "", fmt.Errorf("ifst: TextMarshaler err: %v", err)
		}
		return string(key), nil
	}

	return "", &transferInvalidMapKeyTypeError{Type: vType}
}

func mapToInterfaceTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	inputType := input.Type()

	// todo: Cached
	// todo: How to create reflect of interface{} without TypeOf([]interface{}{})?
	outputType := reflect.MapOf(inputType.Key(), reflect.TypeOf([]interface{}{}).Elem())

	newMap := reflect.MakeMap(outputType)
	err := mapToMapTransfer(input, newMap, opt)
	if err != nil {
		return err
	}

	output.Set(newMap)
	return nil
}

func mapToMapTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	inputType := input.Type()
	outputType := output.Type()

	err := mapKeyValidate(inputType.Key())
	if err != nil {
		return err
	}

	// todo: transfer key too?
	if inputType.Key().Kind() != outputType.Key().Kind() {
		return &transferTypeError{Input: inputType, Output: outputType}
	}

	if output.IsNil() {
		output.Set(reflect.MakeMap(outputType))
	}

	keys := input.MapKeys()
	for _, k := range keys {
		valueType := outputType.Elem()
		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
		}

		outputElem := reflect.New(valueType)
		err := reflectValueTransfer(input.MapIndex(k), outputElem, opt)
		if err != nil {
			return err
		}

		if outputType.Elem().Kind() == reflect.Ptr {
			output.SetMapIndex(k, outputElem)
		} else {
			output.SetMapIndex(k, outputElem.Elem())
		}
	}

	return nil
}

func mapToStructTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	inputType := input.Type()
	// outputType := output.Type()

	err := mapKeyValidate(inputType.Key())
	if err != nil {
		return err
	}

	fieldMap, err := getStructFieldMap(output, opt.Transformer.TagName)
	if err != nil {
		return err
	}

	keys := input.MapKeys()
	for _, k := range keys {
		strKey, err := mapKeyMarshalText(k)
		if err != nil {
			return err
		}

		of, exists := fieldMap[strKey]
		if !exists || !of.IsValid() {
			continue
		}

		if of.Kind() == reflect.Ptr && of.IsNil() {
			of.Set(reflect.New(of.Type().Elem()))
		}

		err = reflectValueTransfer(input.MapIndex(k), of, opt)
		if err != nil {
			return err
		}
	}

	return nil
}

func structToInterfaceTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	outputType := reflect.TypeOf(map[string]interface{}{})

	newMap := reflect.MakeMap(outputType)
	err := structToMapTransfer(input, newMap, opt)
	if err != nil {
		return err
	}

	output.Set(newMap)
	return nil
}

func structToMapTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	inputType := input.Type()
	outputType := output.Type()

	if outputType.Key().Kind() != reflect.String {
		return &transferInvalidMapKeyTypeError{Type: outputType.Key()}
	}

	if output.IsNil() {
		output.Set(reflect.MakeMap(outputType))
	}

	for i := 0; i < inputType.NumField(); i++ {
		inputField := input.Field(i)
		inputTypeField := inputType.Field(i)

		fieldInfo, err := getStructField(inputTypeField, opt.Transformer.TagName)
		if err != nil {
			return err
		}

		if fieldInfo.Name == "-" {
			continue // Ignore "-" tag name
		}

		if len(fieldInfo.PkgPath) != 0 {
			continue // Ignore unexported field
		}

		if fieldInfo.TagOptions.Omitempty && reflect.DeepEqual(inputField.Interface(), reflect.Zero(inputTypeField.Type).Interface()) {
			continue // Ignore zero value
		}

		if fieldInfo.TagOptions.Inline {
			err := reflectValueTransfer(inputField, output, opt)
			if err != nil {
				return nil
			}
			continue
		}

		valueType := outputType.Elem()
		if valueType.Kind() == reflect.Ptr {
			valueType = valueType.Elem()
		}
		outputElem := reflect.New(valueType)
		err = reflectValueTransfer(inputField, outputElem, opt)
		if err != nil {
			return err
		}

		if outputType.Elem().Kind() == reflect.Ptr {
			output.SetMapIndex(reflect.ValueOf(fieldInfo.Name), outputElem)
		} else {
			output.SetMapIndex(reflect.ValueOf(fieldInfo.Name), outputElem.Elem())
		}

	}

	return nil
}

func structToStructTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	inputType := input.Type()
	// outputType := output.Type()

	fieldMap, err := getStructFieldMap(output, opt.Transformer.TagName)
	if err != nil {
		return err
	}

	for i := 0; i < inputType.NumField(); i++ {
		inputField := input.Field(i)
		inputTypeField := inputType.Field(i)

		fieldInfo, err := getStructField(inputTypeField, opt.Transformer.TagName)
		if err != nil {
			return err
		}

		if fieldInfo.Name == "-" {
			continue // Ignore "-" tag name
		}

		if len(fieldInfo.PkgPath) != 0 {
			continue // Ignore unexported field
		}

		if fieldInfo.TagOptions.Omitempty && reflect.DeepEqual(inputField.Interface(), reflect.Zero(inputTypeField.Type).Interface()) {
			continue // Ignore zero value
		}

		if fieldInfo.TagOptions.Inline {
			err := reflectValueTransfer(inputField, output, opt)
			if err != nil {
				return nil
			}
			continue
		}

		of, exists := fieldMap[fieldInfo.Name]
		if !exists || !of.IsValid() {
			continue
		}

		if of.Kind() == reflect.Ptr && of.IsNil() {
			of.Set(reflect.New(of.Type().Elem()))
		}

		err = reflectValueTransfer(inputField, of, opt)
		if err != nil {
			return err
		}
	}

	return nil
}

func structToArrayTransfer(input reflect.Value, output reflect.Value, opt *transferOptions) error {
	inputType := input.Type()
	outputType := output.Type()

	in := inputType.NumField()

	if outputType.Kind() == reflect.Slice {
		if output.Cap() < in {
			newSlice := reflect.MakeSlice(output.Type(), 0, in)
			output.Set(newSlice)
		}
		output.SetLen(in)
	}

	for i := 0; i < in && i < output.Len(); i++ {
		inputField := input.Field(i)
		inputTypeField := inputType.Field(i)

		fieldInfo, err := getStructField(inputTypeField, opt.Transformer.TagName)
		if err != nil {
			return err
		}

		if fieldInfo.Name == "-" {
			continue // Ignore "-" tag name
		}

		if len(fieldInfo.PkgPath) != 0 {
			continue // Ignore unexported field
		}

		if fieldInfo.TagOptions.Omitempty && reflect.DeepEqual(inputField.Interface(), reflect.Zero(inputTypeField.Type).Interface()) {
			continue // Ignore zero value
		}

		if fieldInfo.TagOptions.Inline {
			err := reflectValueTransfer(inputField, output, opt)
			if err != nil {
				return nil
			}
			continue
		}

		err = reflectValueTransfer(inputField, output.Index(i), opt)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	typeTransferTable = map[reflect.Kind]map[reflect.Kind]transferFunc{
		reflect.Bool: {
			reflect.Interface: literalToInterfaceLiteralTransfer,
			reflect.Bool:      literalToInterfaceLiteralTransfer,
		},

		reflect.String: {
			reflect.Interface: literalToInterfaceLiteralTransfer,
			reflect.String:    stringToStringTransfer,
			reflect.Int64:     stringToIntTransfer,
			reflect.Uint64:    stringToUintTransfer,
			reflect.Float64:   stringToFloatTransfer,
		},

		// integer type upgraded to Int64
		reflect.Int64: {
			reflect.Interface: literalToInterfaceLiteralTransfer,
			reflect.Int64:     intToIntTransfer,
			reflect.Uint64:    intToUintTransfer,
			reflect.Float64:   intToFloatTransfer,
			reflect.String:    intToStringTransfer,
		},

		// unsigned integer type upgraded to Uint64
		reflect.Uint64: {
			reflect.Interface: literalToInterfaceLiteralTransfer,
			reflect.Int64:     uintToIntTransfer,
			reflect.Uint64:    uintToUintTransfer,
			reflect.Float64:   uintToFloatTransfer,
			reflect.String:    uintToStringTransfer,
		},

		// float type upgraded to Float64
		reflect.Float64: {
			reflect.Interface: literalToInterfaceLiteralTransfer,
			reflect.Float64:   floatToFloatTransfer,
			reflect.Int64:     floatToIntTransfer,
			reflect.Uint64:    floatToUintTransfer,
			reflect.String:    floatToStringTransfer,
		},

		reflect.Slice: {
			reflect.Interface: sliceToInterfaceTransfer,
			reflect.Slice:     sliceToSliceTransfer,
			reflect.Struct:    arrayToStructTransfer,
			reflect.Array:     arrayToArrayTransfer,
			// todo: transfer into map?
		},

		reflect.Array: {
			reflect.Interface: sliceToInterfaceTransfer,
			reflect.Array:     arrayToArrayTransfer,
			reflect.Struct:    arrayToStructTransfer,
			reflect.Slice:     sliceToSliceTransfer,
			// todo: transfer into map?
		},

		reflect.Interface: {
			reflect.Invalid: interfaceTransfer, // reflect.Invalid represent all type
		},

		reflect.Ptr: {
			reflect.Invalid: ptrTransfer, // reflect.Invalid represent all type
		},

		reflect.Map: {
			reflect.Interface: mapToInterfaceTransfer,
			reflect.Map:       mapToMapTransfer,
			reflect.Struct:    mapToStructTransfer,
			// todo: transfer into Slice or Array?
		},

		reflect.Struct: {
			reflect.Interface: structToInterfaceTransfer,
			reflect.Map:       structToMapTransfer,
			reflect.Struct:    structToStructTransfer,
			reflect.Slice:     structToArrayTransfer,
			reflect.Array:     structToArrayTransfer,
		},
	}
}
