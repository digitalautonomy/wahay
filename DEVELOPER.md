# Code Standards

### Comments
Go provides C-style /* */ block comments and C++-style // line comments. Line comments are the norm; block comments appear mostly as package comments, but are useful within an expression or to disable large swaths of code. 

The program—and web server—godoc processes Go source files to extract documentation about the contents of the package. Comments that appear before top-level declarations, with no intervening newlines, are extracted along with the declaration to serve as explanatory text for the item. The nature and style of these comments determines the quality of the documentation godoc produces. 
```go
$ go doc -all
```

### Names
Names are as important in Go as in any other language. They even have semantic effect: **the visibility of a name outside a package is determined by whether its first character is upper case.** It's therefore worth spending a little time talking about naming conventions in Go programs. 

A good name must be:
* Consistent (easy to guess),
* Short (easy to type),
* Accurate (easy to understand),
* Names in Go should use MixedCase,
* Don't use names_with_underscores,
* Acronyms should be all capitals, as in ServeHTTP and IDProcessor

### Local Variables
Keep them short; long names obscure what the code does. Common variable/type combinations may use really short names:

- Prefer i to index.
- Prefer r to reader.
- Prefer b to buffer.

Avoid redundant names, given their context:
 
Longer names may help in long functions, or functions with many local variables. 

Example bad practice:
```go
func RuneCount(buffer []byte) int {

    runeCount := 0
    for index := 0; index < len(buffer); {
        if buffer[index] < RuneSelf {
            index++
        } else {
            _, size := DecodeRune(buffer[index:])
            index += size
        }
        runeCount++
    }
    return runeCount
}
```
Example good practice:
```go
func RuneCount(b []byte) int {
    count := 0
    for i := 0; i < len(b); {
        if b[i] < RuneSelf {
            i++
        } else {
            _, n := DecodeRune(b[i:])
            i += n
        }
        count++
    }
    return count
}
```

### Packages
By convention, packages are given lower case, single-word names; there should be no need for underscores or mixedCaps.

Another convention is that the package name is the base name of its source directory; the package in src/encoding/base64 is imported as "encoding/base64" but has name base64, not encoding_base64 and not encodingBase64. 

The package name is only the default name for imports; it need not be unique across all source code, and in the rare case of a collision the importing package can choose a different name to use locally. In any case, confusion is rare because the file name in the import determines just which package is being used. 

Choose package names that lend meaning to the names they export.

Steer clear of util, common, and the like. 

### Exported package-level names
Exported names are qualified by their package names.

Remember this when naming exported variables, constants, functions, and types.

That's why we have bytes.Buffer and strings.Reader,
not bytes.ByteBuffer and strings.StringReader. 

### Parameters
Function parameters are like local variables,
but they also serve as documentation.

Where the types are descriptive, they should be short: 

```go
func AfterFunc(d Duration, f func()) *Timer

func Escape(w io.Writer, s []byte)
```

Where the types are more ambiguous, the names may provide documentation: 
```go
func Unix(sec, nsec int64) Time

func HasPrefix(s, prefix []byte) bool
```

### Return values
Return values on exported functions should only be named for documentation purposes.

These are good examples of named return values: 
```go
func Copy(dst Writer, src Reader) (written int64, err error)

func ScanBytes(data []byte, atEOF bool) (advance int, token []byte, err error)
```

### Method Receivers
Receivers are a special kind of argument.

By convention, they are one or two characters that reflect the receiver type,
because they typically appear on almost every line: 
```go
func (b *Buffer) Read(p []byte) (n int, err error)

func (sh serverHandler) ServeHTTP(rw ResponseWriter, req *Request)

func (r Rectangle) Size() Point
```
Receiver names should be consistent across a type's methods.
(Don't use r in one method and rdr in another.)


### Interface Types
Interfaces that specify just one method are usually just that function name with 'er' appended to it. 
```go
type Reader interface {
    Read(p []byte) (n int, err error)
} 
```
Sometimes the result isn't correct English, but we do it anyway: 
```go
type Execer interface {
    Exec(query string, args []Value) (Result, error)
}
```

Sometimes we use English to make it nicer: 
```go
type ByteReader interface {
    ReadByte() (c byte, err error)
}
```

When an interface includes multiple methods, choose a name that accurately describes its purpose (examples: net.Conn, http.ResponseWriter, io.ReadWriter). 

### Errors
Error types should be of the form FooError: 
```go
type ExitError struct {
    ...
}
```

Error values should be of the form ErrFoo: 
```go
var ErrFormat = errors.New("image: unknown format")
```

### Import paths
The last component of a package path should be the same as the package name. 
```go
"compress/gzip" // package gzip
```

Avoid stutter in repository and package paths: 
```go
"code.google.com/p/goauth2/oauth2" // bad; my fault
```

For libraries, it often works to put the package code in the repo root:
```go
"github.com/golang/oauth2" // package oauth2
```

Also avoid upper case letters (not all file systems are case sensitive). 

### Getters
Go doesn't provide automatic support for getters and setters. There's nothing wrong with providing getters and setters yourself, and it's often appropriate to do so, but it's neither idiomatic nor necessary to put Get into the getter's name. If you have a field called owner (lower case, unexported), the getter method should be called Owner (upper case, exported), not GetOwner. The use of upper-case names for export provides the hook to discriminate the field from the method. A setter function, if needed, will likely be called SetOwner. Both names read well in practice: 
```go
owner := obj.Owner()
if owner != user {
    obj.SetOwner(user)
}
```


### Tips
Techniques to write Go code that is
* simple,
* readable,
* maintainable,
* handling error using ```return``` statement
```go
func (g *Gopher) WriteTo(w io.Writer) (size int64, err error) {
    err = binary.Write(w, binary.LittleEndian, int32(len(g.Name)))
    if err != nil {
        return
    }
    size += 4
    n, err := w.Write([]byte(g.Name))
    size += int64(n)
    if err != nil {
        return
    }
    err = binary.Write(w, binary.LittleEndian, int64(g.AgeYears))
    if err == nil {
        size += 4
    }
    return
}
```

Deploy one-off utility types for simpler code 
```go
type binWriter struct {
    w    io.Writer
    size int64
    err  error
}
```

```go
// Write writes a value to the provided writer in little endian form ((least significant value in the sequence) is stored first).
func (w *binWriter) Write(v interface{}) {
    if w.err != nil {
        return
    }
    if w.err = binary.Write(w.w, binary.LittleEndian, v); w.err == nil {
        w.size += int64(binary.Size(v))
    }
}
```

* Type switch to handle special cases
```go
func (w *binWriter) Write(v interface{}) {
    if w.err != nil {
        return
    }
    switch v.(type) {
    case string:
        s := v.(string)
        w.Write(int32(len(s)))
        w.Write([]byte(s))
    case int:
        i := v.(int)
        w.Write(int64(i))
    default:
        if w.err = binary.Write(w.w, binary.LittleEndian, v); w.err == nil {
            w.size += int64(binary.Size(v))
        }
    }
}
```

* Type switch with short variable declaration
```go
func (w *binWriter) Write(v interface{}) {
    if w.err != nil {
        return
    }
    switch x := v.(type) {
    case string:
        w.Write(int32(len(x)))
        w.Write([]byte(x))
    case int:
        w.Write(int64(x))
    default:
        if w.err = binary.Write(w.w, binary.LittleEndian, v); w.err == nil {
            w.size += int64(binary.Size(v))
        }
    }
}
```

* Shorter is better (or at least longer is not always better)

* Try to find the shortest name that is self explanatory
```go
Prefer MarshalIndent to MarshalWithIndentation
```

* Don't forget that the package name will appear before the identifier you chose
```go
In package encoding/json we find the type Encoder, not JSONEncoder.
It is referred as json.Encoder.
```

* Avoid very long files
* Separate code and tests
* Separated package documentation (When we have more than one file in a package, it's convention to create a doc.go
containing the package documentation)
* Make your packages "go get"-able (Some packages are potentially reusable, some others are not)
* Ask for what you need

Let's use the Gopher type:
```go
type Gopher struct {
    Name     string
    AgeYears int
}
```
We could define this method 
```go
func (g *Gopher) WriteToFile(f *os.File) (int64, error) {
```
But using a concrete type makes this code difficult to test, so we use an interface. 
```go
func (g *Gopher) WriteToReadWriter(rw io.ReadWriter) (int64, error) {
```
And, since we're using an interface, we should ask only for the methods we need. 
```go
func (g *Gopher) WriteToWriter(f io.Writer) (int64, error) {
```

* Keep packages independent
```go
import (
    "golang.org/x/talks/content/2013/bestpractices/funcdraw/drawer"
    "golang.org/x/talks/content/2013/bestpractices/funcdraw/parser"
)
```

* Avoid dependency by using an interface
```go
import "image"

// Function represent a drawable mathematical function.
type Function interface {
    Eval(float64) float64
}

// Draw draws an image showing a rendering of the passed Function.
func Draw(f Function) image.Image {
```

* Using an interface instead of a concrete type makes testing easier. 


### References
* Andrew Gerrand (adg@golang.org), Google Inc. (October 2014), What's in a name? (https://talks.golang.org/2014/names.slide)
* Francesc Campoy Flores (gopher@google.com), Google Inc, Twelve Go Best Practices (https://talks.golang.org/2013/bestpractices.slide)
