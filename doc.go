package morph

import (
	"encoding/json"
	"slices"
	"strings"
)

type DocFunctionRenderer interface {
	RenderJSON(functions *FunctionStore) []byte
}

// public type to facilitate reading and parsing function information, for the purpose of building doc files
func NewFunctionDocs(fs *FunctionStore) *FunctionDocs {
	ret := &FunctionDocs{}
	ret.buildSections(fs)
	return ret
}

type FunctionDocs struct {
	Sections []*docFnNamespace `json:"sections"`
}

type docFnNamespace struct {
	Name             string          `'json:"name"`
	FunctionSections []*docFnSection `json:"function_sections"`
}

type docFnSection struct {
	Tag       string        `json:"tag"`
	Namespace string        `json:"namespace"`
	Functions []*docFnEntry `json:"functions"`
}

type docFnEntry struct {
	Name        string              `json:"name"`
	Namespace   string              `json:"namespace"`
	Signature   string              `json:"signature"`
	Description string              `json:"description"`
	Tags        []FunctionTag       `json:"tags"`
	Args        []FunctionArg       `json:"args"`
	Return      *FunctionReturn     `json:"return"`
	Attributes  []FunctionAttribute `json:"attributes"`
	Examples    []ProgramExample    `json:"examples"`
}

func (d *FunctionDocs) RenderJSON(functions *FunctionStore) ([]byte, error) {
	return json.Marshal(d)
}

var orderedTagList = []FunctionTag{
	FUNCTION_TAG_GENERAL,
	FUNCTION_TAG_ERR_NULL_CHECKS,
	FUNCTION_TAG_TYPE_COERCION,
	FUNCTION_TAG_FLOW_CONTROL,
	FUNCTION_TAG_NUMBERS,
	FUNCTION_TAG_STRINGS,
	FUNCTION_TAG_ARRAYS,
	FUNCTION_TAG_MAPS,
	FUNCTION_TAG_TIME,
	FUNCTION_TAG_HIGHER_ORDER,
}

func (d *FunctionDocs) buildSections(fs *FunctionStore) {
	//list namespaces by starting with "std", but then use alphabetical order after
	nsNameList := []string{}
	for nsName := range fs.namespaces {
		if nsName == "std" {
			continue
		}
		nsNameList = append(nsNameList, nsName)
	}
	slices.Sort(nsNameList)
	nsNameList = append([]string{"std"}, nsNameList...)

	// now we can iterate through in order
	for _, nsName := range nsNameList {
		ns := fs.namespaces[nsName]
		nsToAdd := &docFnNamespace{
			Name:             nsName,
			FunctionSections: []*docFnSection{},
		}
		alreadyIncluded := []string{}
		for _, tag := range orderedTagList {

			fnSectionToAdd := &docFnSection{
				Tag:       string(tag),
				Namespace: ns.name,
				Functions: []*docFnEntry{},
			}

			for _, fnEntry := range ns.functions {
				if !slices.Contains(fnEntry.Tags, tag) {
					continue
				}
				if slices.Contains(alreadyIncluded, fnEntry.Name) {
					continue
				}
				fnToAdd := &docFnEntry{
					Name:        fnEntry.Name,
					Namespace:   fnEntry.namespace,
					Signature:   fnEntry.Signature(),
					Description: fnEntry.Description,
					Tags:        fnEntry.Tags,
					Args:        fnEntry.Args,
					Attributes:  fnEntry.Attributes,
					Examples:    fnEntry.Examples,
				}
				if fnEntry.Return != nil {
					fnToAdd.Return = fnEntry.Return
				}
				fnSectionToAdd.Functions = append(fnSectionToAdd.Functions, fnToAdd)
				alreadyIncluded = append(alreadyIncluded, fnEntry.Name)
			}
			slices.SortFunc(fnSectionToAdd.Functions, func(a *docFnEntry, b *docFnEntry) int {
				return strings.Compare(a.Name, b.Name)
			})
			if len(fnSectionToAdd.Functions) > 0 {
				nsToAdd.FunctionSections = append(nsToAdd.FunctionSections, fnSectionToAdd)
			}
		}
		d.Sections = append(d.Sections, nsToAdd)
	}
}
