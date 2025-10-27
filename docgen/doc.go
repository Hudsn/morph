package docgen

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/hudsn/morph"
)

type DocFunctionRenderer interface {
	RenderJSON(functions *morph.FunctionStore) []byte
}

// public type to facilitate reading and parsing function information, for the purpose of building doc files
func NewFunctionDocs(fs *morph.FunctionStore) *FunctionDocs {
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
	Name        string                    `json:"name"`
	Namespace   string                    `json:"namespace"`
	Signature   string                    `json:"signature"`
	Description string                    `json:"description"`
	Tags        []morph.FunctionTag       `json:"tags"`
	Args        []morph.FunctionArg       `json:"args"`
	Return      *morph.FunctionReturn     `json:"return"`
	Attributes  []morph.FunctionAttribute `json:"attributes"`
	Examples    []morph.ProgramExample    `json:"examples"`
}

func (dfn *docFnEntry) JoinedTags() string {
	tagList := []string{}
	for _, entry := range dfn.Tags {
		tagList = append(tagList, string(entry))
	}
	return strings.Join(tagList, ", ")
}

func (dfn *docFnEntry) FormatReturn() string {
	if dfn.Return == nil {
		return "No usable or assignable object is returned from this function"
	}
	return dfn.Return.Description
}

func (d *FunctionDocs) RenderJSON(functions *morph.FunctionStore) ([]byte, error) {
	return json.Marshal(d)
}

var orderedTagList = []morph.FunctionTag{
	morph.FUNCTION_TAG_GENERAL,
	morph.FUNCTION_TAG_ERR_NULL_CHECKS,
	morph.FUNCTION_TAG_TYPE_COERCION,
	morph.FUNCTION_TAG_FLOW_CONTROL,
	morph.FUNCTION_TAG_NUMBERS,
	morph.FUNCTION_TAG_STRINGS,
	morph.FUNCTION_TAG_ARRAYS,
	morph.FUNCTION_TAG_MAPS,
	morph.FUNCTION_TAG_TIME,
	morph.FUNCTION_TAG_HIGHER_ORDER,
}

func (d *FunctionDocs) buildSections(fs *morph.FunctionStore) {
	//list namespaces by starting with "std", but then use alphabetical order after
	nsNameList := []string{}
	for nsName := range fs.Namespaces {
		if nsName == "std" {
			continue
		}
		nsNameList = append(nsNameList, nsName)
	}
	slices.Sort(nsNameList)
	nsNameList = append([]string{"std"}, nsNameList...)

	// now we can iterate through in order
	for _, nsName := range nsNameList {
		ns := fs.Namespaces[nsName]
		nsToAdd := &docFnNamespace{
			Name:             nsName,
			FunctionSections: []*docFnSection{},
		}
		alreadyIncluded := []string{}
		for _, tag := range orderedTagList {

			fnSectionToAdd := &docFnSection{
				Tag:       string(tag),
				Namespace: ns.Name,
				Functions: []*docFnEntry{},
			}

			for _, fnEntry := range ns.Functions {
				if !slices.Contains(fnEntry.Tags, tag) {
					continue
				}
				if slices.Contains(alreadyIncluded, fnEntry.Name) {
					continue
				}
				fnToAdd := &docFnEntry{
					Name:        fnEntry.Name,
					Namespace:   fnEntry.Namespace,
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
