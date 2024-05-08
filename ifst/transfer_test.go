package ifst_test

import (
	"encoding/json"
	"fmt"
	"hago-plat/pcom/ifst"
	"reflect"
	"testing"

	"gopkg.in/mgo.v2/bson"
)

type MapKey struct {
	Key string
}

func (this *MapKey) MarshalText() (text []byte, err error) {
	return []byte(this.Key), nil
}

func TestBoolTransfer(t *testing.T) {
	input := true
	var output interface{}
	expect := input

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestIntTransfer(t *testing.T) {
	input := int(999)
	var output interface{}
	expect := input

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestStringTransfer(t *testing.T) {
	input := "hello"
	var output interface{}
	expect := input

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestFloatTransfer(t *testing.T) {
	input := float32(0.5)
	var output interface{}
	expect := input

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestInvalidTypeCase(t *testing.T) {
	input := int(999)
	var output bool

	err := ifst.Transfer(input, &output)
	if err == nil {
		t.Error("invalid type case check fail")
		return
	}
	t.Log("Success: error message ->", err)
}

func TestInvalidTypeCase2(t *testing.T) {
	input := float32(0.5)
	var output int

	err := ifst.Transfer(input, &output)
	if err == nil {
		t.Error("invalid type case check fail")
		return
	}
	t.Log("Success: error message ->", err)
}

func TestArrayToInterfaceTransfer(t *testing.T) {
	input := [5]int{1, 2, 3, 4, 5}
	var output interface{}
	expect := []interface{}{1, 2, 3, 4, 5}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestSliceToInterfaceTransfer2(t *testing.T) {
	input := []struct {
		A int
	}{
		{1},
		{2},
	}

	var output interface{}

	expect := []interface{}{
		map[string]interface{}{
			"A": 1,
		},
		map[string]interface{}{
			"A": 2,
		},
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestArrayLessLengthTransfer(t *testing.T) {
	input := [5]float32{1.1, 2.2, 3.3, 4.4, 5.5}
	var output [4]float32
	expect := [4]float32{1.1, 2.2, 3.3, 4.4}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestArrayGreaterLengthTransfer(t *testing.T) {
	input := [5]string{"a", "b", "c", "d", "e"}

	output := [6]string{}
	output[5] = "hello world!"

	expect := [6]string{"a", "b", "c", "d", "e", ""}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestArrayTransferInvalidCase(t *testing.T) {
	input := [5]string{"a", "b", "c", "d", "e"}
	var output [5]int

	err := ifst.Transfer(input, &output)
	if err == nil {
		t.Error("invalid type case check fail")
		return
	}
	t.Log("Success: error message ->", err)
}

func TestArrayToStructTransfer(t *testing.T) {
	input := []interface{}{
		1,
		"hello",
		0.5,
		true,
	}

	type O struct {
		A int
		B string
		C float64
		D bool
	}

	var output O
	expect := O{1, "hello", 0.5, true}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestSliceToInterfaceTransfer(t *testing.T) {
	input := []string{"a", "b", "c", "d", "e"}
	var output interface{}
	expect := []interface{}{"a", "b", "c", "d", "e"}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestSliceTransfer(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	var output []int
	expect := input

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestSliceTransferInvalidCase(t *testing.T) {
	input := []string{"a", "b", "c", "d", "e"}
	var output []int

	err := ifst.Transfer(input, &output)
	if err == nil {
		t.Error("invalid type case check fail")
		return
	}
	t.Log("Success: error message ->", err)
}

func TestSliceGreaterLengthTransfer(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	output := make([]int, 0, 10)
	expect := input

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestMapToInterface(t *testing.T) {
	input := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	var output interface{}

	expect := map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestMapToInterface2(t *testing.T) {
	ka := &MapKey{"a"}
	kb := &MapKey{"b"}
	kc := &MapKey{"c"}

	input := map[*MapKey]int{
		ka: 1,
		kb: 2,
		kc: 3,
	}

	var output interface{}
	expect := map[*MapKey]interface{}{
		ka: 1,
		kb: 2,
		kc: 3,
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestMapToMapTransfer(t *testing.T) {
	input := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	var output map[string]int

	expect := input

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestMapToMapTransferInvalidCase(t *testing.T) {
	input := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	var output map[string]float32

	err := ifst.Transfer(input, &output)
	if err == nil {
		t.Error("invalid type case check fail")
		return
	}
	t.Log("Success: error message ->", err)
}

func TestMapToMapTransferInvalidCase2(t *testing.T) {
	input := map[int64]int{
		1: 1,
		2: 2,
		3: 3,
	}

	var output map[int]int

	err := ifst.Transfer(input, &output)
	if err == nil {
		t.Error("invalid type case check fail")
		return
	}
	t.Log("Success: error message ->", err)
}

func TestMapToStructTransfer(t *testing.T) {
	input := map[string]int{
		"A": 1,
		"B": 2,
	}

	type O struct {
		A int
		B int
	}

	var output O
	expect := O{1, 2}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestMapToEmbeddedStructTransfer(t *testing.T) {
	input := map[string]interface{}{
		"A": 1,
		"B": 2,
		"C": map[string]int{
			"AA": 3,
		},
		"D": 4,
	}

	type EmbAnonymous struct {
		D int
	}

	type Emb struct {
		AA int
	}

	type O struct {
		A            int
		B            int
		C            Emb
		EmbAnonymous `ifst:",inline"`
	}

	var output O
	expect := O{1, 2, Emb{3}, EmbAnonymous{4}}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestMapToEmbeddedStructTransfer2(t *testing.T) {
	input := map[string]interface{}{
		"A": 1,
		"B": 2,
		"C": map[string]int{
			"AA": 3,
		},
		"EmbAnonymous": map[string]int{
			"D": 4,
		},
	}

	type EmbAnonymous struct {
		D int
	}

	type Emb struct {
		AA int
	}

	type O struct {
		A int
		B int
		C Emb
		EmbAnonymous
	}
	var output O
	expect := O{1, 2, Emb{3}, EmbAnonymous{4}}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestMapToEmbeddedPtrStructTransfer(t *testing.T) {
	input := map[string]interface{}{
		"A": 1,
		"B": 2,
		"C": map[string]int{
			"AA": 3,
		},
	}

	type Emb struct {
		AA int
	}

	type O struct {
		A int
		B *int
		C *Emb
	}

	var output O

	eb := 2
	expect := O{1, &eb, &Emb{3}}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestTextMarshalerTransfer(t *testing.T) {
	ka := &MapKey{"A"}
	kb := &MapKey{"B"}
	kc := &MapKey{"C"}
	kaa := &MapKey{"AA"}

	input := map[*MapKey]interface{}{
		ka: 1,
		kb: 2,
		kc: map[*MapKey]int{
			kaa: 3,
		},
	}

	type Emb struct {
		AA int
	}

	type O struct {
		A int
		B int
		C Emb
	}

	var output O
	expect := O{1, 2, Emb{3}}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestTextMarshalerTransferInvalidCase(t *testing.T) {

	input := map[MapKey]interface{}{
		MapKey{"A"}: 1,
		MapKey{"B"}: 2,
		MapKey{"C"}: map[MapKey]int{
			MapKey{"AA"}: 3,
		},
	}

	type Emb struct {
		AA int
	}

	output := struct {
		A int
		B int
		C Emb
	}{}

	err := ifst.Transfer(input, &output)
	if err == nil {
		t.Error("invalid type case check fail")
		return
	}
	t.Log("Success: error message ->", err)
}

type StructToMapBase struct {
	I int64
}

func TestStructToMapTransfer(t *testing.T) {
	input := struct {
		A  int
		M  map[string]StructToMapBase
		MP map[string]*StructToMapBase
	}{A: 1,
		M: map[string]StructToMapBase{
			"a": {I: 2},
		},
		MP: map[string]*StructToMapBase{
			"b": {I: 999},
		}}

	output := map[string]interface{}{}
	expect := map[string]interface{}{
		"A": 1,
		"M": map[string]interface{}{
			"a": map[string]interface{}{
				"I": int64(2),
			},
		},
		"MP": map[string]interface{}{
			"b": map[string]interface{}{
				"I": int64(999),
			},
		},
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestStructToMapTransfer2(t *testing.T) {
	input := struct {
		M  StructToMapBase
		MP *StructToMapBase
	}{
		M: StructToMapBase{
			I: 2,
		},
		MP: &StructToMapBase{
			I: 999,
		}}

	output := map[string]interface{}{}
	expect := map[string]interface{}{
		"M": map[string]interface{}{
			"I": int64(2),
		},
		"MP": map[string]interface{}{
			"I": int64(999),
		},
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestStructToMapTransfer3(t *testing.T) {
	input := struct {
		M  StructToMapBase
		MP *StructToMapBase
	}{
		M: StructToMapBase{
			I: 2,
		},
		MP: &StructToMapBase{
			I: 999,
		}}

	output := map[string]StructToMapBase{}
	expect := map[string]StructToMapBase{
		"M": {
			I: 2,
		},
		"MP": {
			I: 999,
		},
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestStructToMapTransfer4(t *testing.T) {
	input := struct {
		M  StructToMapBase
		MP *StructToMapBase
	}{
		M: StructToMapBase{
			I: 2,
		},
		MP: &StructToMapBase{
			I: 999,
		}}

	output := map[string]*StructToMapBase{}
	expect := map[string]*StructToMapBase{
		"M": {
			I: 2,
		},
		"MP": {
			I: 999,
		},
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestStructToInterfaceTransfer(t *testing.T) {
	type EmbAnonymous struct {
		C int
	}

	type EmbAnonymousInline struct {
		D int
	}

	type Emb struct {
		A int
	}

	type I struct {
		A int
		B Emb
		EmbAnonymous
		EmbAnonymousInline `ifst:",inline"`
	}

	input := I{1, Emb{2}, EmbAnonymous{3}, EmbAnonymousInline{4}}

	var output interface{}
	expect := map[string]interface{}{
		"A": 1,
		"B": map[string]interface{}{
			"A": 2,
		},
		"EmbAnonymous": map[string]interface{}{
			"C": 3,
		},
		"D": 4,
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestStructToStructTransfer(t *testing.T) {
	input := struct {
		A int
		B string
		C float32
		D bool
	}{1, "hello", 0.5, true}

	type O struct {
		A int64
		B string
		C float64
		D bool
	}

	var output O
	expect := O{1, "hello", 0.5, true}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

type StructToStructBase struct {
	Int int64
}

func TestStructToStructTag(t *testing.T) {
	input := struct {
		A  int    `ifst:"C"`
		B  string `ifst:"D"`
		a  int
		M  map[string]StructToStructBase
		MP map[string]*StructToStructBase
	}{A: 1,
		B: "hello",
		a: 2,
		M: map[string]StructToStructBase{
			"tt": {Int: 66},
		}, MP: map[string]*StructToStructBase{
			"tt": {Int: 66},
		}}

	var output struct {
		AA int    `ifst:"C"`
		BB string `ifst:"D"`
		a  int
		M  map[string]StructToStructBase
		MP map[string]*StructToStructBase
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if input.A != output.AA || input.B != output.BB {
		t.Errorf("unexpected value: %+v vs %+v", input, output)
		return
	}
}

func TestStructToSliceTransfer(t *testing.T) {
	input := struct {
		A int
		B string
		C float64
		D bool
	}{1, "hello", 0.5, true}

	output := make([]interface{}, 0, 4)
	expect := []interface{}{1, "hello", 0.5, true}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestStructToArrayTransfer(t *testing.T) {
	input := struct {
		A int
		B string
		C float64
		D bool
	}{1, "hello", 0.5, true}

	output := [3]interface{}{}
	expect := [3]interface{}{1, "hello", 0.5}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestNilPointerInputInvalidCase(t *testing.T) {
	var input *int
	var output int

	err := ifst.Transfer(input, &output)
	if err == nil {
		t.Error("invalid type case check fail")
		return
	}
	t.Log("Success: error message ->", err)
}

func TestNilPointerOutputInvalidCase(t *testing.T) {
	input := int(999)
	var output *int

	err := ifst.Transfer(input, output)
	if err == nil {
		t.Error("invalid type case check fail")
		return
	}
	t.Log("Success: error message ->", err)
}

func TestInterfaceInvalidCase(t *testing.T) {
	var input interface{}
	var output int

	err := ifst.Transfer(input, &output)
	if err == nil {
		t.Error("invalid type case check fail")
		return
	}
	t.Log("Success: error message ->", err)
}

func TestInterfaceInvalidCase2(t *testing.T) {
	var input interface{}
	var output int

	err := ifst.Transfer(&input, &output)
	if err != nil {
		t.Error(err)
		return
	}
}

type UnmarshalerStruct struct {
	A int
	B string
}

func (this *UnmarshalerStruct) UnmarshalIFST(input interface{}) error {
	switch input.(type) {
	case int:
		this.A = input.(int)
	case string:
		this.B = input.(string)
	default:
		return fmt.Errorf("invalid type of input: %v", reflect.ValueOf(input).Kind())
	}

	return nil
}

func TestNormalToUnmarshalerTransfer(t *testing.T) {
	input := int(999)
	var output UnmarshalerStruct
	expect := UnmarshalerStruct{
		A: input,
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

type MarshalerStruct struct {
	A string
}

func (this *MarshalerStruct) MarshalIFST() (interface{}, error) {
	return this.A, nil
}

func TestMarshalerToNormalTransfer(t *testing.T) {
	input := &MarshalerStruct{
		A: "hello",
	}
	var output string
	expect := input.A

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestMarshalerToUnmarshalerTransfer(t *testing.T) {
	input := &MarshalerStruct{
		A: "hello",
	}
	var output UnmarshalerStruct
	expect := UnmarshalerStruct{
		B: input.A,
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestInlineTransfer(t *testing.T) {
	type Emb struct {
		A int
	}
	type I struct {
		Data Emb `ifst:",inline"`
	}

	input := I{
		Data: Emb{
			A: 99,
		},
	}
	output := map[string]interface{}{}
	expect := map[string]interface{}{
		"A": 99,
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestInlineTransfer2(t *testing.T) {
	type Emb struct {
		A int
	}
	type I struct {
		Data Emb
	}

	input := I{
		Data: Emb{
			A: 99,
		},
	}
	output := map[string]interface{}{}
	expect := map[string]interface{}{
		"Data": map[string]interface{}{
			"A": 99,
		},
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestInlineTransfer3(t *testing.T) {
	type Emb struct {
		A int
	}
	type O struct {
		Data Emb `ifst:",inline"`
	}

	input := map[string]interface{}{
		"A": 99,
	}
	output := O{}
	expect := O{
		Data: Emb{
			A: 99,
		},
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestInlineTransfer4(t *testing.T) {
	type Emb struct {
		A int
	}
	type O struct {
		Data Emb
	}

	input := map[string]interface{}{
		"Data": map[string]interface{}{
			"A": 99,
		},
	}
	output := O{}
	expect := O{
		Data: Emb{
			A: 99,
		},
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestOmitemptyTransfer(t *testing.T) {
	type I struct {
		A  int `ifst:",omitempty"`
		AA int `ifst:",omitempty"`

		B  string `ifst:",omitempty"`
		BB string `ifst:",omitempty"`

		C  float64 `ifst:",omitempty"`
		CC float64 `ifst:",omitempty"`

		D  map[string]string `ifst:",omitempty"`
		DD map[string]string `ifst:",omitempty"`

		E  *int `ifst:",omitempty"`
		EE *int `ifst:",omitempty"`

		F  []string `ifst:",omitempty"`
		FF []string `ifst:",omitempty"`
	}

	ee := 10
	input := I{
		AA: 999,
		BB: "Hello",
		CC: 5.5,
		DD: map[string]string{
			"key": "",
		},
		EE: &ee,
		FF: []string{
			"Good",
		},
	}
	output := map[string]interface{}{}
	expect := map[string]interface{}{
		"AA": 999,
		"BB": "Hello",
		"CC": 5.5,
		"DD": map[string]interface{}{
			"key": "",
		},
		"EE": 10,
		"FF": []interface{}{
			"Good",
		},
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestOmitemptyTransfer2(t *testing.T) {
	type O struct {
		A  int `ifst:",omitempty"`
		AA int `ifst:",omitempty"`

		B  string `ifst:",omitempty"`
		BB string `ifst:",omitempty"`

		C  float64 `ifst:",omitempty"`
		CC float64 `ifst:",omitempty"`

		D  map[string]string `ifst:",omitempty"`
		DD map[string]string `ifst:",omitempty"`

		E  *int `ifst:",omitempty"`
		EE *int `ifst:",omitempty"`

		F  []string `ifst:",omitempty"`
		FF []string `ifst:",omitempty"`
	}

	ee := 10
	input := map[string]interface{}{
		"AA": 999,
		"BB": "Hello",
		"CC": 5.5,
		"DD": map[string]interface{}{
			"key": "",
		},
		"EE": 10,
		"FF": []interface{}{
			"Good",
		},
	}
	output := O{}
	expect := O{
		AA: 999,
		BB: "Hello",
		CC: 5.5,
		DD: map[string]string{
			"key": "",
		},
		EE: &ee,
		FF: []string{
			"Good",
		},
	}

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
	if output.E != nil {
		t.Errorf("unexpected output(%v): expect nil point", reflect.TypeOf(output).String())
		return
	}
}

func TestTypenameOfLiteral(t *testing.T) {
	type O string
	input := "hello"
	var output O
	expect := input

	err := ifst.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if expect != string(output) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

type JsonMarshaler struct {
	hhh string `json:"a"`
}

func (this *JsonMarshaler) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.hhh)
}

func TestJsonMarshaler(t *testing.T) {

	input := JsonMarshaler{
		hhh: "hello",
	}
	var output string
	expect := "hello"

	err := ifst.TransferWithTagName(&input, &output, "json")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

type JsonMarshaler2 struct {
	hhh string `json:"a"`
}

func (this JsonMarshaler2) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.hhh)
}

type JsonMarshalerParent struct {
	M JsonMarshaler2 `json:"m"`
}

func TestJsonMarshaler2(t *testing.T) {

	input := JsonMarshalerParent{
		M: JsonMarshaler2{"hello"},
	}
	var output interface{}
	expect := map[string]interface{}{
		"m": "hello",
	}

	err := ifst.TransferWithTagName(input, &output, "json")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

type JsonUnmarshaler struct {
	hhh string `json:"a"`
}

func (this *JsonUnmarshaler) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &this.hhh)
}

func TestJsonUnmarshaler(t *testing.T) {
	input := "hello"
	var output JsonUnmarshaler
	expect := JsonUnmarshaler{
		hhh: "hello",
	}

	err := ifst.TransferWithTagName(input, &output, "json")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

type JsonUnmarshalerParent struct {
	M JsonUnmarshaler `json:"m"`
}

func TestJsonUnmarshalerParent(t *testing.T) {
	input := map[string]interface{}{
		"m": "hello",
	}
	var output JsonUnmarshalerParent
	expect := JsonUnmarshalerParent{
		M: JsonUnmarshaler{"hello"},
	}

	err := ifst.TransferWithTagName(input, &output, "json")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestJsonMarshalerAndUnmarshaler(t *testing.T) {
	input := JsonMarshaler{
		hhh: "hello",
	}
	var output JsonUnmarshaler
	expect := JsonUnmarshaler{
		hhh: "hello",
	}

	err := ifst.TransferWithTagName(&input, &output, "json")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

type BsonGetter struct {
	hhh string `bson:"a"`
}

func (this BsonGetter) GetBSON() (interface{}, error) {
	return map[string]interface{}{"bson": this.hhh}, nil
}

func TestBsonGetter(t *testing.T) {

	input := BsonGetter{
		hhh: "hello",
	}
	var output map[string]interface{}
	expect := map[string]interface{}{"bson": "hello"}

	err := ifst.TransferWithTagName(&input, &output, "bson")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

type BsonSetter struct {
	hhh string `json:"a"`
}

func (this *BsonSetter) SetBSON(raw bson.Raw) error {
	var v map[string]interface{}
	err := raw.Unmarshal(&v)
	if err != nil {
		fmt.Println("err: ", err)
		return err
	}

	s, exists := v["bson"]
	if !exists {
		return fmt.Errorf("setter field not exists")
	}

	str, ok := s.(string)
	if !ok {
		return fmt.Errorf("invalid typee of setter value")
	}

	this.hhh = str
	return nil
}

func TestBsonSetter(t *testing.T) {
	input := map[string]interface{}{"bson": "hello"}
	var output BsonSetter
	expect := BsonSetter{
		hhh: "hello",
	}

	err := ifst.TransferWithTagName(input, &output, "bson")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestBsonGetterAndSetter(t *testing.T) {
	input := BsonGetter{
		hhh: "hello",
	}
	var output BsonSetter
	expect := BsonSetter{
		hhh: "hello",
	}

	err := ifst.TransferWithTagName(&input, &output, "bson")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}

func TestWeakTypeTransfer(t *testing.T) {
	input := 999
	var output string
	expect := "999"

	tr := ifst.NewTransformer()
	tr.EnableWeakTypeTransfer()

	err := tr.Transfer(input, &output)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(output, expect) {
		t.Errorf("unexpected output(%v): %+v, expect(%v): %+v", reflect.TypeOf(output).String(), output, reflect.TypeOf(expect).String(), expect)
		return
	}
}
