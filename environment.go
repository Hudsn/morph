package parser

type environment struct {
	store map[string]object
	outer *environment
}
