package quintus_test

import (
	"fmt"
	"testing"

	"hago-plat/pcom/quintus"
)

func TestObjectName(t *testing.T) {
	ExportedCamel := quintus.ObjectName("HelloWorld")
	UnexportedCamel := quintus.ObjectName("helloWorld")
	Snake := quintus.ObjectName("hello_world")

	Misc1 := quintus.ObjectName("_hello_world_")
	Misc2 := quintus.ObjectName("hello_1world")
	Misc3 := quintus.ObjectName("1hello_1world")
	Misc4 := quintus.ObjectName("__hello_world__")

	samples := []func() error{
		ValidateSampleFunc(ExportedCamel, ExportedCamel.ExportedCamel(), "HelloWorld"),
		ValidateSampleFunc(ExportedCamel, ExportedCamel.UnexportedCamel(), "helloWorld"),
		ValidateSampleFunc(ExportedCamel, ExportedCamel.Snake(), "hello_world"),

		ValidateSampleFunc(UnexportedCamel, UnexportedCamel.ExportedCamel(), "HelloWorld"),
		ValidateSampleFunc(UnexportedCamel, UnexportedCamel.UnexportedCamel(), "helloWorld"),
		ValidateSampleFunc(UnexportedCamel, UnexportedCamel.Snake(), "hello_world"),

		ValidateSampleFunc(Snake, Snake.ExportedCamel(), "HelloWorld"),
		ValidateSampleFunc(Snake, Snake.UnexportedCamel(), "helloWorld"),
		ValidateSampleFunc(Snake, Snake.Snake(), "hello_world"),

		ValidateSampleFunc(Misc1, Misc1.ExportedCamel(), "_helloWorld_"),
		ValidateSampleFunc(Misc1, Misc1.UnexportedCamel(), "_helloWorld_"),
		ValidateSampleFunc(Misc1, Misc1.Snake(), "_hello_world_"),

		ValidateSampleFunc(Misc2, Misc2.ExportedCamel(), "Hello1World"),
		ValidateSampleFunc(Misc2, Misc2.UnexportedCamel(), "hello1World"),
		ValidateSampleFunc(Misc2, Misc2.Snake(), "hello_1world"),

		ValidateSampleFunc(Misc3, Misc3.ExportedCamel(), "1hello1World"),
		ValidateSampleFunc(Misc3, Misc3.UnexportedCamel(), "1hello1World"),
		ValidateSampleFunc(Misc3, Misc3.Snake(), "1hello_1world"),

		ValidateSampleFunc(Misc4, Misc4.ExportedCamel(), "_HelloWorld_"),
		ValidateSampleFunc(Misc4, Misc4.UnexportedCamel(), "_HelloWorld_"),
		ValidateSampleFunc(Misc4, Misc4.Snake(), "__hello_world__"),
	}

	for i, fn := range samples {
		err := fn()
		if err != nil {
			t.Errorf("fail: idx:%v, %v", i, err)
			return
		}
	}
}

func ValidateSampleFunc(name quintus.ObjectName, expect string, result string) func() error {
	return func() error {
		return ValidateSample(name, expect, result)
	}
}

func ValidateSample(name quintus.ObjectName, result string, expect string) error {
	if expect != result {
		return fmt.Errorf("name:%v, expect:%v, result:%v", name, expect, result)
	}
	return nil
}
