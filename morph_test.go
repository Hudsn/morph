package morph

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMorphBasicExample(t *testing.T) {
	test := testMorphCase{
		description: "basic io test",
		srcJSON: `
		{
			"mood": "happy"
		}
		`,
		in: `
			WHEN src.mood == "happy" :: SET dest = "🙂"
		`,
		wantJSON: `
		"🙂"
		`,
	}
	checkTestMorphCase(t, test)
}

func checkTestMorphCase(t *testing.T, tt testMorphCase) bool {
	m, err := New(tt.in)
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.ToJSON([]byte(tt.srcJSON))
	if err != nil {
		t.Fatal(err)
	}

	var gotInterface interface{}
	err = json.Unmarshal(got, &gotInterface)
	if err != nil {
		t.Fatal(err)
	}
	var wantInteface interface{}
	err = json.Unmarshal([]byte(tt.wantJSON), &wantInteface)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantInteface, gotInterface) {
		t.Errorf("%s: WRONG VALUE\n\twant:\n\t\t%s\n\tgot\n\t\t%s\n", tt.description, tt.wantJSON, string(got))
		return false
	}
	return true
}

type testMorphCase struct {
	description string
	in          string
	srcJSON     string
	wantJSON    string
}
