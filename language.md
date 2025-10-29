# Language Guide

## How it works

A Morph program consists of three major components:
- An input JSON blob, accessible in Morph syntax as: `@in`
- A series of one or more Morph statments
- An output JSON blob, assignable in Morph syntax as: `@out`

To sucessfully transform and return data, you must assign the desired values or fields to the `@out` object via `SET` statments.

For example, setting `@out.my_value` to `5`, will result in a final object that looks like: 

```
{
    "my_value": 5
}
```

## Accessing data items

You can reference any data object via its variable name. 

`@in` will be the only variable available at the start of any program. You can access it directly, or you can use different expression types depending on the data type of the variable you're referencing.

For example, if `@in` (or any other target variable) is an integer, float, string, or boolean, you can only access it directly via its name.

If your target variable is an object with sub-fields, you can access them via `.` path notation, such as `@in.my_field.my_nested_field`

If your target variable is an array, you can reference a specific index with `[int]` notation, such as `myarray[4]` or `myarr[2+2]`

You can also chain these ways of accessing data. For example, if you set a variable that is an object with an array inside it, you can access an index of that array like: `myobj.nested_arr[0]`

## SET Statements
`SET` statments are the only way to create and set variables in Morph. 

The only variable `SET` will not work on is the "@in" variable, which cannot be modified.

A `SET` statement follows the syntax: 

`SET variable = value`

Note that when setting a variable to another variable like `SET x = y`, the right side variable is cloned before being assigned, meaning that future changes to `x` should ***not*** change `y`. 

Note that `SET` is case insensitive, but it is encouraged to use all-caps for readability.

## DEL Statements
`DEL` statements delete a given variable at a target path except "@in", which cannot be modified.

A `DEL` statement follows the syntax: 

`DEL variable`

Note that `DEL` is case insensitive, but it is encouraged to use all-caps for readability.


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
SET @out.text = @in.text // You can also add single line comments like this!
// or like this!
SET @out.emoji = "😶"
IF @in.text == "happy" :: SET @out.emoji = "🙂"
IF @in.text == "sad" :: SET @out.emoji = "☹️"

//out
{
    "text": "happy"
    "emoji": "🙂"
}
```

## Types

Morph has a few `BASIC` types that are available to be used as a part of any expression or statement:

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
    - commonly encountered when referencing variables that don't exist. For example: `@in.doesnt_exist` would return `NULL`
    - always `falsey`
- ERROR
    - an item representing an error that happened during the execution of a morph program.
    - uncaught/unhandled errors will exit the current run of the morph program, and the calling `morph.Exec` will return an error to be handled by your Go application as you see fit.
    - the default builtin funciton store provides basic tools to handle errors thrown by morph expressions

The above types are all included in the `BASIC` and `ANY` function signature types.

### Arrow Functions
Morph also has a special type: the `Arrow Function`.  

Arrow functions are used in higher-order builtin functions like `map()`, `filter()` and `reduce()`, and allow the function caller to invoke a user-defined sub-sequence of Morph statements as part of that function. This is a useful pattern for working with maps and arrays.

Arrow functions are only usable as function arguments where specified, and are **NOT** directly callable in a basic morph statement. 
Any function that specifies a parameter can be of type `ARROW` or `ANY` can use.

The `ARROW` type is included in the `ANY` function signature type.

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
//@in:
{
    "name": "Daniel"
    "cool_factor": 999
}

// morph
SET is_cool = @in.cool_factor >= 500
SET @out.name = @in.name
IF @in.name == "Daniel" || is_cool :: SET @out.name = 'The Cooler ${@in.name}'


//@out:
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

Morph exposes multiple builtin functions (with more to be added) to handle common transformation tasks.

These functions have their own dedicated page that is automatically generated from the source repository.

You can access the builtin function docs [here](https://hudsn.github.io/morph/). 

### Custom Functions 

To make the language extendable, you can create custom functions and add them to the function store that powers each instance of morph. 

You can see an example on how to do this in the `lang` directory, or by clicking [here](lang/customize.md)