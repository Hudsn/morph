# Custom Functions

To make the language extendable, you can create custom functions and add them to the function store that powers each instance of morph. 
So if you think any of the existing builtins are hot garbage, or you simply want to build your own specific functionality, you can build it!

An example is probably the easiest way to explain how to do this.

First, you'll need to initialize a function store. 
There is a public function to initialize one with the existing builtins.

```go
myFuncStore := morph.DefaultFunctionStore()
// you can also use NewFunctionStore() if you don't want any of the existing builtins
```

Then we'll need to create our custom function.
Let's make something that doubles the `INTEGER` passed to it, and returns an `INTEGER`

Custom functions follow the signature
```go
func(ctx context.Context, args ...*lang.Object) *lang.Object
```

where `Object` represents a general structure in morph such that it can be any of the underlying types dicussed in the language guide. It might be easier to think of it as an `ANY` type.

The `Object` type exposes functions that deal with converting back and forth beween native Go types and Objects.

Let's try it out. Here's our function

```go
func myIntDoubler(ctx context.Context, args ...*lang.Object) *lang.Object {
    // here we enforce that there should only be a single argument
    if res, ok := lang.IsArgCountEqual(1, args); !ok {
        // in this case res will be an internal lang.ERROR type which is automatically genrated by the IsArgCountEqual function when the specified arg counts don't match the actual arg count
        return res 
    }  

    // then cast the first argument as an integer (int64)
    num, err := args[0].AsInt()
    if err != nil {
        return ObjectError(err.Error()) 
        // this will return a generic morph conversion error on function call if the first argument is not an integer.
        // but you can pass whatever message you want here
    }
    return CastInt(num * 2) 
    // this does what it says: casts the passed value as a morph Integer.
    // If a Go type is given that is not convertable, it will give a generic morph conversion error that should bubble up to where you're running morph from.
    // If you want to return a custom error, you can simply assign the results of CastInt to appropriate variables rather than returning the CastInt call directly.
}

```


then you'll need to create a custom function entry and register it. There are a functions for that too:
```go
myFuncEntry := lang.NewFunctionEntry(
    "my_func", // the name of the function (how it will be called in morph); note that having entries with the same name in the same namespace will cause one to overwrite the other
    "Doubles the passed integer", // description of the function,
    myIntDoubler, // the go function that satisfies the morph signature
    lang.WithArgs(
        lang.NewFunctionArg(
				"num",
				"the number we'll be doubling",
				lang.INTEGER,
        ),
    ),
    lang.WithReturn(
        lang.NewFunctionReturn(
            "the doubled result",
            lang.INTEGER
        ),
    ),
    WithExamples( // we can even give examples that can be used for documentation generation
        NewProgramExample(
            `{"input": 2}`
            `SET out.result = my_func(@in.input)`
            `{"result": 4}`
        ),
    ),
)

myFuncStore.Register(myFuncEntry) // registers to std namespace
// you can also register functions to a custom namespace so that the function needs to be called like: my_custom_namespace.my_func(2)
myFuncStore.RegisterToNamespace("my_custom_namespace", myFuncEntry) 
```

Now we can initialize morph with our store: 

```go
jsonIn := []byte(`{"number": 2}`)
// we're useing both both the custom namespace and the std one, for demonstration purposes:
input := `
SET from_custom = my_custom_namespace.my_func(@in.number)
SET @out.result = my_func(from_custom)
`
//initialize morph with our custom function store
m, err := morph.New(input, WithFunctionStore(myFuncStore))
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
```json
{"result": 8}
```

That's it! We built and registered our custom function and it should work!

**Fun fact:** All of the "builtin" functions are actually implemented this way in`builtin.go`, so check it out if you need a reference or example!