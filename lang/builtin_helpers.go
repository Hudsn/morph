package lang

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
)

// iterate through all namespaces in a function store, and run each of their registered functions to ensure they produce the expected result.
func RunFunctionStoreExamples(fs *FunctionStore) error {
	nsList := []string{}
	for _, ns := range fs.Namespaces {
		nsList = append(nsList, ns.Name)
	}
	slices.Sort(nsList)
	for _, nsName := range nsList {
		ns := fs.Namespaces[nsName]
		fnList := []string{}
		for _, fn := range ns.Functions {
			fnList = append(fnList, fn.Name)
		}
		slices.Sort(fnList)

		for _, fnName := range fnList {
			fn := ns.Functions[fnName]
			for idx, exEntry := range fn.Examples {
				err := testBuiltinFunctionEntry(fs, exEntry.In, exEntry.Program, exEntry.Out)
				if err != nil {
					return fmt.Errorf("error running example #%d with %s.%s: %w", idx+1, ns.Name, fn.Name, err)
				}
			}
		}
	}
	return nil
}

func testBuiltinFunctionEntry(fstore *FunctionStore, in string, programContents string, out string) error {
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

	runnableProgram, err := NewProgram(programContents, fstore)
	if err != nil {
		return err
	}

	gotJSON, err := runnableProgram.Run([]byte(in))
	if err != nil {
		return err
	}

	// gotCast, ok := gotAny.(map[string]interface{})
	// if !ok {
	// 	log.Fatal("askdjfkasjfd")
	// }
	// arr, ok := gotCast["result"]
	// if !ok {
	// 	log.Fatal("lkkjdfakjfkajfdk")
	// }
	// arrCast, ok := arr.([]interface{})
	// if !ok {
	// 	log.Fatal("kjasdfkjasfafd")
	// }
	// for _, entry := range arrCast {
	// 	fmt.Printf("GOT:: %T :: %#v\n", entry, entry)
	// }

	// wantCast, ok := want.(map[string]interface{})
	// if !ok {
	// 	log.Fatal("askdjfkasjfd")
	// }
	// arr, ok = wantCast["result"]
	// if !ok {
	// 	log.Fatal("lkkjdfakjfkajfdk")
	// }
	// arrCast, ok = arr.([]interface{})
	// if !ok {
	// 	log.Fatal("kjasdfkjasfafd")
	// }
	// for _, entry := range arrCast {
	// 	fmt.Printf("WANT:: %T :: %#v\n", entry, entry)
	// }

	var got interface{}
	err = json.Unmarshal([]byte(gotJSON), &got)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(want, got) {

		return fmt.Errorf("wrong value for input:\n%s\n\nwant:\n\t%s\ngot\n\t%s", programContents, out, string(gotJSON))
	}

	return nil
}
