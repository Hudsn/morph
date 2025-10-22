package morph

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"testing"
)

func TestBuiltins(t *testing.T) {
	err := testBuiltinStore(newBuiltinFunctionStore())
	if err != nil {
		t.Fatal(err)
	}
}

func testBuiltinStore(fs *FunctionStore) error {
	nsList := []string{}
	for _, ns := range fs.namespaces {
		nsList = append(nsList, ns.name)
	}
	slices.Sort(nsList)
	for _, nsName := range nsList {
		ns := fs.namespaces[nsName]
		fnList := []string{}
		for _, fn := range ns.functions {
			fnList = append(fnList, fn.Name)
		}
		slices.Sort(fnList)

		for _, fnName := range fnList {
			fn := ns.functions[fnName]
			for idx, exEntry := range fn.Examples {
				err := testBuiltinFE(fs, exEntry.In, exEntry.Program, exEntry.Out)
				if err != nil {
					return fmt.Errorf("error testing example #%d with %s.%s: %w", idx+1, ns.name, fn.Name, err)
				}
			}
		}
	}
	return nil
}

func testBuiltinFE(fstore *FunctionStore, in string, programContents string, out string) error {
	if len(in) != 0 {
		var inputValidator interface{}
		err := json.Unmarshal([]byte(in), &inputValidator)
		if err != nil {
			return fmt.Errorf("invalid input json: %w", err)
		}
	}

	var want interface{}
	err := json.Unmarshal([]byte(out), &want)
	if err != nil {
		return err
	}
	runnableProgram, err := New(programContents, WithFunctionStore(fstore))
	if err != nil {
		return err
	}

	gotJSON, err := runnableProgram.ToJSON([]byte(in))
	if err != nil {
		return err
	}
	var got interface{}
	err = json.Unmarshal([]byte(gotJSON), &got)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(want, got) {

		return fmt.Errorf("wrong value for input:\n%s\n\nwant:\n\t%s\ngot\n\t%s\n", programContents, out, string(gotJSON))
	}

	return nil
}
