# Morph
A zero-dependency JSON transformation language for Go.

## About

### What even is this?

This is a domain-specific language (DSL) that allows users to transform and re-shape JSON data.

A simple program looks something like this: 
```
// input:
{
    "name": "Fluffy"
}

// Morph program:
SET dest.name == src.name
SET dest.pet_type = "unknown"
WHEN src.name == "Fluffy" :: SET dest.pet_type = "dog"

// output:
{
    "name": "Fluffy",
    "pet_type": "dog"
}
```

### When would I use it?
You can use this anywhere you'd want to expose arbitrary or user-defined transformations for JSON data in your Go application. 

One use case could be to use this language as one part of a larger configuration-driven data pipeline that is backed by Go. 

For example: Your users could define transformations using Morph in a `.yaml` file, which would then be read and executed by your Go application.

### Motivation

I wanted a way to allow users to define custom transformations in configuration files. While there were existing libraries, and even entire frameworks for data transformation, I was unable to find one fit all of my requirements:
 
 - Low number of dependencies (ideally none)
 - No reliance on external services
 - Usable/embeddable in a configuration file, and runnable in Go
 - Shared transformation and expression language

Also, learning and edification.

### Design goals 

- **Simple**: Should be a small keyword footprint with easy to understand syntax and minimal nesting. There should be one fairly obvious way to do a given thing.

- **Intuitive**: Should cater to the easiest and most common transformation use cases, while offering opt-in ways to achieve more complex mappings and transforms.

- **Extendable**: In cases where the baseline features that aim to satisfy the "simple" and "intuitive" requirements are insufficient, a library API should allow the language to be extended to achieve the desired functionality.

## Quick Start

Install:

```go get github.com/hudsn/morph```


Run a minimal program:

```go
jsonIn := []byte(`{"hello": "world"}`)
input := `
SET dest.message = "hello " + src.hello
`
m, err := morph.New(input)
if err != nil {
    log.Fatal(err)
}
jsonBytes, err := m.ToJSON(jsonIn)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(jsonBytes))
```

outputs:
```
{
    "message": "hello world" 
}
```

## Further Usage

If you want to learn more about the language, including how to register and use your own custom functions, a more detailed [language guide](language.md) is available. 


## Contributing

Pull requests and issues are welcome.

If submitting a pull request, please ensure that you have created relevant tests in the appropriate `_test.go` file with a similar naming convention to other tests in that file. For example in `lexer_test.go`, all tests start with `TestLex...`. This lets us run all tests of a given type. Using the lexer example, all lexer tests can be run with `go test -run "TestLex"`

You can run the entire test suite like any standard Go project `go test ./...`.

If adding a builtin function, please make a best effort to add the function to the appropriate section in the `language.md` file and register it to the default namespace.