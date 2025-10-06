# Language Guide

## How it works

A Morph program consists of three major components:
- A source JSON blob, accessible in Morph syntax as: `src`
- A series of one or more Morph statments
- A destination JSON blob, assignable in Morph syntax as: `dest`

To sucessfully transform and return data, you must assign the desired values or fields to the `dest` object via `SET` statments.

For example, setting `dest.my_value` to `5`, will result in a final object that looks like: 

```
{
    "my_value": 5
}
```

## Accessing data items

You can reference any data object via its variable name. 

`src` will be the only variable available at the start of any program. You can access it directly, or you can use different expression types depending on the data type of the variable you're referencing.

For example, if `src` (or any other target variable) is an integer, float, string, or boolean, you can only access it directly via its name.

If your target variable is an object with sub-fields, you can access them via `.` path notation, such as `src.my_field.my_nested_field`

If your target variable is an array, you can reference a specific index with `[int]` notation, such as `myarray[4]` or `myarr[2+2]`

You can also combine these. For example, if you set a variable that is an object with an array inside it, you can access an index of that array like: `myobj.nested_arr[0]`

## SET Statements
`SET` statments are the only way to create and set variables in Morph. 

A `SET` statement follows the syntax: 

`SET variable = value`

Note that when setting a variable to another variable like `SET x = y`, the right side variable is cloned, meaning that future changes to `x` should ***not*** change `y`. 

Note that `SET` is case insensitive, but it is encouraged to use all-caps for readability.


## IF Statements
`IF` statements are the only way to conditionally execute other statements, namely `SET` statements.

`IF` statements can be single line statements or spam multiple lines

A single-lined `IF` statement follows the syntax: 

`IF condition :: consequence`

Single line `IF` statements **MUST** have a consequence of a SET statement.

A mult-iline `IF` statement follows the syntax:

```
IF condition :: {
    consequence statement 1
    consequence statement 2
    ...
}
```

Multi-line `IF` statements **MUST** have its consequence statement(s) enclosed in curly brackets as shown above

If the condition evalutes to be `true`, the consequence will execute.

Note that `IF` is case insensitive, but it is encouraged to use all-caps for readability.


## Example
Let's say we want to output an "emoji" field based on the input's "text" field, and also preserve the "text" field:
- If the text is "happy", we'll output a 🙂. 
- If the text is "sad", we'll output a ☹️.
- Otherwise, we'll output a 😶.

Here is what that would look like:

```
//in 
{
    "text": "happy"
}

//morph program:
SET dest.text = src.text // You can also add single line comments like this!
// or like this!
SET dest.emoji = "😶"
IF src.text == "happy" :: SET dest.emoji = "🙂"
IF src.text == "sad" :: SET dest.emoji = "☹️"

//out
{
    "text": "happy"
    "emoji": "🙂"
}
```

## Types

Morph has 8 main types that are available to be used as a part of any expression or statement:

- Boolean
    -  `true` or `false`
- String
    - any value captured between two `"` characters. For example: `"Hello world!"`
    - you can also use `'` instead of double-quotes to declare multi-line strings
    - `'` strings can also make use of template strings `${}` to format expressions into strings. For example: `'my ${1300 + 37} ${"str" + "ing"}'` results in `my 1337 string`
    - all values are `truthey` **except for empty strings `""`**

- Integer
    - any whole 64-bit number, negatives allowed. For example: `999` or `-999`
    - all values are`truthey` **except for `0`** 
- Float
    - any floating-point 64-bit number, negatives allowed. For example. `999.999` or `-999.999`
    - all values are`truthey` **except for `0`** 
- Array
    - a comma-separated list of values enclosed between square braces. For example: `[1, 2, "three"]`
    - arrays can be of mixed types
    - all values are `truthey` **except for empty arrays `[]`**  
- Map
    - a collection of key:value pairs, expressed as a comma-separated list of pairs between curly braces. For example: `{"key": "value", "hello": "world"}`
    - keys **MUST** be strings
    - values can be any of these main Morph types
    - all values are `truthey` **except for empty maps `{}`**
- Time
    - an item representing a timestamp
    - must be explicity delcared or parsed from a string using functions; JSON typically uses strings or integers to represent time.
    - all values are `truthey` **except for the 'zero' time value equal to January 1, year 1, 00:00:00 UTC** 
- NULL
    - a non-value; empty, like my soul when writing documentation, expressed as a keyword `NULL` or `null`
    - commonly encountered when referencing variables that don't exist. For example: `src.doesnt_exist` would return `NULL`
    - always `falsey`



Note: There is also an **internal** `Error` type. When an uncaught error (see builtin `catch` or `fallback` functions for how to catch them) is encountered by Morph, the program will stop, and the calling `morph.ToJSON` or `morph.ToAny` will return an error to be handled by your Go application as you see fit.


## Operators

Morph supports multiple operators to write expressions

### Prefix
`!` can be used before any boolean expression to return the opposite of that value. For example`!true` would result in `false`.

`-` can be used before any number to respresent its negative value. For example: `-999`

### Comparison

Morph supports equality and inequality checks that you are probably already familiar with:
- `<` less than
    - numbers only
- `<=` greater than or equal to
    - numbers only
- `>` greater than
    - numbers only
- `>=` greater than or equal to
    - numbers only
- `==` equal
    - numbers, booleans, or strings
- `!=` not equal
    - numbers, booleans, or strings

Note that these operators do not work on Arrays or Maps

### Logical
- `&&` logical AND
- `||` logical OR

### Numbers

- `+` add two numbers
- `-` subtract the right number from the left number
- `*` multiply two numbers
- `/` divide the left number by the right number
- `%` divide the left number by the right number, and return the remainder

### Strings
- `+` concatenate two strings

## The Cooler Example

We've learned a bit more, so let's have another example using some of the operators and expressions.

```
// src:
{
    "name": "Daniel"
    "cool_factor": 999
}

// morph
SET is_cool = src.cool_factor >= 500
SET dest.name = src.name
IF src.name == "Daniel" || is_cool :: SET dest.name = 'The Cooler ${src.name}'


//dest:
{
    "name": "The Cooler Daniel"
}
```

## Functions

Morph uses functions to perform slightly more complex operations.

By design, you cannot define functions directly in Morph.
However, you **CAN** make use of various built-in functions, or even define your own custom Morph functions using Go (more on that later).

You can call functions with the following notation: `function_name(argument1, argument2)` or if there are no arguments: `my_function()`

Functions always return a single value, which can be any of the main types. Usage might look like this: `SET my_variable = my_function()`

If the function is called incorrectly, it will return an error instead of the intended type. Don't worry, Morph exposes builtin functions to handle this case. Read the next sections to see how...

### Namespaces

Functions can be called based on the namespace they're registered to via `.` path notation. For example `mycoolnamespace.mycoolfunction()`.
Note that this path notation is **NOT** chainable, and is limited to a single `.` path operator.

By default, all builtin functions are registered in the `std` namespace, which is a special namespace whose functions can be called without referencing the namespace. For example `std.myfunc()` can simply be called via `myfunc()`

You can register custom functions to any custom namespace, or to the `std` namespace. See the Custom Functions section for more information.

## Builtin Functions

Morph exposes multiple builtin functions (with more to be added) to handle common transformation tasks:

### catch(item, fallback): result
Checks if `item` is an error, and if so, returns the `fallback` value instead. Otherwise, returns `item`

Example: 
```
SET my_variable =  catch(non_existent_func(), "my fallback!")
// my_variable results in "my fallback" because 
```
### coalesce(item, fallback): result
Checks if `item` is `NULL`, and if so, returns the `fallback` value instead. Otherwise, returns `item`

Example: 
```
SET my_variable =  coalesce(NULL, "coalesce ftw!")
// my_variable results in "coalesce"
```

### fallback(item, fallback): result
A combination of `catch` and `coalesce`. Checks if `item` is `NULL` or an error, and if so, returns the `fallback` value instead. Otherwise, returns `item`.

### drop()
Stops the current run of the Morph program, and returns dest as NULL.
Useful in IF statements for short-circuiting

### emit()
Stops the current run of the Morph program, and returns dest at its current state.
Useful in IF statements for short-circuiting

Example:
```
SET dest = "hello world"
emit()
SET dest = "goodbye world"

// returns "hello world"
```

### int(item): int
Converts the target item to an int.
Argument type must be `INTEGER`, `FLOAT`, or `STRING`

### float(item): string
Converts the target item to a float.
Argument type must be `INTEGER`, `FLOAT`, or `STRING`

### string(item): string
Converts the target item to a string.
Argument type must be `INTEGER`, `FLOAT`, `STRING`, or `BOOLEAN`

### len(item): int
Returns the length of the passed `STRING` or `ARRAY` as an `INTEGER`.

### min(num1, num2): num
Returns the smaller of two numbers, which can be a combination of `INT` or `FLOAT` 

### max(num1, num2): num
Returns the larger of two numbers, which can be a combination of `INT` or `FLOAT` 

### contains(item1, item2): bool
Returns `true` if item2 is contained within item1, otherwise returns false.
First argument must be an `ARRAY` or `STRING`.
Second argument can be any type if the first argument is an `ARRAY`, but must be a `STRING` if the first argument is a `STRING`

### append(array, item) array
Returns the first argument `ARRAY` with the item added to the end.


### now() time 
Returns a time item representing the current time in UTC.


### A note on arrow function arguments
Some functions take Arrow Functions (`ARROWFUNC`) as an argument type.

This is a special type of argument that is typically used as a way to execute higher order functions.

An arrow function will effectively run its own instance of a morph program with its own starting variable. It has no reference to outer variables.

Here's an example of what it looks like in the `map()` builtin:

```
map(src.my_array, entry ~> {
    SET return = entry.value.item
})
```

In this example, the array `src.my_array` is being made available via the `entry` variable in the arrow function. 

The statements in the curly brackets will run for EACH item in `src.my_array`, which can be accessed inside the brackets as part of the `entry` variable. In the case of `map()`, it's `entry.value`, but the point is that `entry` is the inner variable that will be used to access your initial data. 

You can ignore the `return` variable for now, since it's specific to the `map` function.

### map(input_data, arrow_func): output_data
Applies arrow function statements to the input data. The returned value from the arrow function replaces the original entry.

input_data must be an `ARRAY` or `MAP`


#### For arrays:
The index is accessible at the injected variable's `variable.index` path, and the element itself is available at the injected variable's `variable.value` path.

The desired output element's value should be assigned to the `return` variable in the arrow function

Example: 
```
//input
{
    "my_arr": [1, 2.5, 3]
}

//morph
SET dest.new_arr = map(src.my_arr, entry ~> {
    IF entry.index == 2 :: SET return = entry.value * 2
}) 

//output
{
    "new_arr": [1, 2.5, 6]
}
```

**Note:** 
- If return is unset or `drop()` is used in the arrow function, the original value of that element will be returned.



#### For Maps: 
The key `STRING` is accessible at the injected variable's `variable.key` path, and the element itself is available at the injected variable's `variable.value` path.

The desired output key should be assigned to the `return.key` variable in the arrow function.
The desired output value should be assigned to the `return.value` variable in the arrow function.



Example: 
```
//input: 
{
    "a": 1,
    "b": 2,
    "c": 3
}

//morph:
SET dest = map(src, entry ~> {
    SET return.key = "prefix_" + entry.key 
    SET return.value = entry.value * 2
})


//output:
{
    "prefix_a": 2,
    "prefix_b": 4,
    "prefix_c": 6
}   
```
**Note:** 
- If you assign `return.key` a key that exists in another element in the original map, the original key will be kept to prevent unpredictable and erroneous assignment 
- If `return.key` is unset, the original key is used.
- If `return.value` is unset, the original value is used.
- If `drop()` is called, both original values are kept.

## filter(input_data, arrow_func): output_data
Applies arrow function statements to the input data. If the `return` value from the arrow functions is `true`, the entry is kept. If it is `false` or unset, the entry is removed.

### For arrays:
The index is accessible at the injected variable's `variable.index` path, and the element itself is available at the injected variable's `variable.value` path.

Example:
```
// input:
{
    "my_arr": [1, 2, "three", "4", 4]
}

// morph:
SET dest = filter(src.my_arr, entry ~> {
			IF entry.index >= 2 && (catch(entry.value % 2 == 0, false) || catch(int(entry.value) >= 4, false)) :: SET return = true
    })

//output: 
["4", 4]
```

### For maps: 
The key `STRING` is accessible at the injected variable's `variable.key` path, and the element itself is available at the injected variable's `variable.value` path.

Example:
```
//input:
{
    "a": 1,
    "b": 2,
    "c": 3
}

//morph:
SET res = filter(src, entry ~> {
        IF entry.key == "a" :: SET return = true
        IF entry.value == 3 :: SET return = true
    })

//output:
{
    "a": 1,
    "c": 3,
}

```

## reduce(input_data, starting_accumulator, arrow_func): output_acc
Applies arrow function statements to the input data. Returns the final state of the accumulator.

The `return` value will be the new accumulator value for the next item.

If the `return` value is unset or `drop()` is called, the accumulator will remain the same for that iteration.

For both arrays and maps, the accumulator's current value can be accessed at the injected variable's `variable.current` path. 


### For arrays:
The index is accessible at the injected variable's `variable.index` path, and the element itself is available at the injected variable's `variable.value` path.

Example: 
```
//input:
{
    "my_arr": [1, 2, "3"]
}

//morph:
SET dest.result = reduce(src.my_arr, null, entry ~> {
        IF entry.current == NULL :: SET entry.current = 0
        SET return = entry.current + int(entry.value)
    })

//output:
{
    "result": 6
}
```

### For maps: 
The key `STRING` is accessible at the injected variable's `variable.key` path, and the element itself is available at the injected variable's `variable.value` path.

Example: 
```
//input:
{
    "a": 1,
    "b": 2,
    "c": 3
}

//morph:
SET dest.result = reduce(src, null, entry ~> {
    IF entry.current == NULL :: SET entry.current = 0
    IF entry.key != "a" :: SET return = entry.current + int(entry.value)
})


//output:
{
    "result": 5
}
```

### More builtins coming soon!
More builtin functions are planned.

### Pipes

When calling Functions, you can also use pipe operator `|` syntax. 
This passes the result of the expression on the left side of the pipe as the first argument of the function on the right side of the pipe.

For example: `1 + 2 | min(4)` would translate to `min(3, 4)`

Keep in mind that operator precedence matters here: 

Pipes are **LOWER** precedence than typical arithmetic operators like  `+`, `-`, `*`, `/`, and `%`. So piping the result of those operators works. 

However, the pipe operator is higher precedence than (in)equality operators and boolean operators, so something like this `true == "pizza" | contains("iz")`, would evaluate to true, rather than false (for mismatched comparison types) since the `"pizza" | ...` expression is evaluated first  

### Operator Precedence

Precedence for common operators is as follows (lowest to highest):
- OR `||`
- AND `&&`
- (IN)EQUALITY OPERATORS: `==`, `!=`, `>`, `>=`, etc...
- PIPE `|` 
- ADD/SUBTRACT `+`, `-`
- PRODUCT `*`, `/`, `%`
- PREFIX: `!`, `-(int)` 

### Custom Functions 

To make the language extendable, you can create custom functions and add them to the function store that powers each instance of morph. 

So if you think any of the existing builtins are garbage, or you simply want to build your own specific functionality, you can build it!

An example is probably the easiest way to explain how to do this.

First, you'll need to initialize a function store. 
There is a public function to initialize one with the existing builtins.

```go
myFuncStore := morph.NewDefaultFunctionStore()
// you can also use NewEmptyFunctionStore() if you don't want any of the existing builtins
```

Then we'll need to create our custom function.
Let's make something that doubles the `INTEGER` passed to it, and returns an `INTEGER`

Custom functions follow the signature
```go
func(args ...*morph.Object) *morph.Object
```

where `Object` represents a general structure in morph such that it can be any of the underlying types we've discussed so far. It might be easier to think of it as an `ANY` type.

The `Object` type exposes functions that deal with converting back and forth beween native Go types and Objects.

Let's try it out. Here's our function

```go
func myIntDoubler(args ...*morph.Object) *morph.Object {
    // here we enforce that there should only be a single argument
    if res, ok := morph.IsArgCountEqual(1, args); !ok {
        return res // in this case res will be a morph ERROR which is auto-generated for us by morph
    }  

    // then cast the first argument as an integer (int64)
    num, err := args[0].AsInt()
    if err != nil {
        return ObjectError(err.Error()) 
        // this will return a generic morph conversion error on function call if the first argument is not an integer.
        // you can pass whatever message you want here though...
    }
    return CastInt(num * 2) 
    // this does what it says: casts the passed value as a morph Integer.
    // If a Go type is given that is not convertable, it will give a generic morph conversion error that should bubble up to where you're running morph from.
}

```


then you'll need to create a custom function entry and register it. There are a functions for that too:
```go
myFuncEntry := NewFunctionEntry("my_func")
myFuncEntry.SetArgument("num", INTEGER)
myFuncEntry.SetReturn("result", INTEGER)
myFuncStore.Register(myFuncEntry) // registers to std
// you can also register functions to a custom namespace
myFuncStore.RegisterToNamespace("my_custom_namespace", myFuncEntry) 
```

Now we can initialize morph with our store: 

```go
jsonIn := []byte(`{"number": 2}`)
// we use both instances of the function:
// the custom namespace and the std one, for demonstration purposes:
input := `
SET from_custom = my_custom_namespace.my_func(src.number)
SET dest = my_func(from_custom)
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
```
8
```

That's it! We built and registered our custom function and it works!

**Fun fact:** All of the "builtin" functions are actually implemented this way in`builtin.go`