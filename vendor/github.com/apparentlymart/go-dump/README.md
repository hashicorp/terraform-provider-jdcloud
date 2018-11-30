# go-dump

`dump` is a Go package for generating pretty-printed dumps of Go values.

There are many functions in the Go ecosystem to do this, including one built
in to Go itself in the `fmt` package, with the `%#v` format verb.

This package attempts to find a nice compromise between the built-in formatter
and more advanced formatters like [go-spew](https://github.com/davecgh/go-spew).
Go types often implement the `fmt.GoStringer` interface to produce a more
concise representation of values in Go syntax, but conventionally this result
is a single-line string and thus hard to read for larger data structures.
`go-spew` instead produces a formatted dump of a value by using the `reflect`
package to analyze its contents, but this often exposes the internals of
data structures that make the result hard to read.

`dump.Value` works by first obtaining the `GoString` result for the given
value and then pretty-printing the result so that nested struct, map and
slice literals are easier to read. This allows the result of an overridden
`GoString` implementation to be included while still producing a readable
result.

## Usage

```
go get -u github.com/apparentlymart/go-dump/...
```

```go
t.Logf(dump.Value(v))
```
