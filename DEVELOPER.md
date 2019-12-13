# Code Standards <sup>[1](#tookFrome)</sup>

### Comments
Go provides C-style /* */ block comments and C++-style // line comments. Line comments are the norm; block comments appear mostly as package comments, but are useful within an expression or to disable large swaths of code. 

The program—and web server—godoc processes Go source files to extract documentation about the contents of the package. Comments that appear before top-level declarations, with no intervening newlines, are extracted along with the declaration to serve as explanatory text for the item. The nature and style of these comments determines the quality of the documentation godoc produces. 
```
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
```
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
```
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

### Parameters
Function parameters are like local variables,
but they also serve as documentation.

Where the types are descriptive, they should be short: 

```
func AfterFunc(d Duration, f func()) *Timer

func Escape(w io.Writer, s []byte)
```

Where the types are more ambiguous, the names may provide documentation: 
```
func Unix(sec, nsec int64) Time

func HasPrefix(s, prefix []byte) bool
```

### Return values
Return values on exported functions should only be named for documentation purposes.

These are good examples of named return values: 
```
func Copy(dst Writer, src Reader) (written int64, err error)

func ScanBytes(data []byte, atEOF bool) (advance int, token []byte, err error)
```

### Receivers
Receivers are a special kind of argument.

By convention, they are one or two characters that reflect the receiver type,
because they typically appear on almost every line: 
```
func (b *Buffer) Read(p []byte) (n int, err error)

func (sh serverHandler) ServeHTTP(rw ResponseWriter, req *Request)

func (r Rectangle) Size() Point
```
Receiver names should be consistent across a type's methods.
(Don't use r in one method and rdr in another.)

### Exported package-level names
Exported names are qualified by their package names.

Remember this when naming exported variables, constants, functions, and types.

That's why we have bytes.Buffer and strings.Reader,
not bytes.ByteBuffer and strings.StringReader. 

### Interface Types
Interfaces that specify just one method are usually just that function name with 'er' appended to it. 
```
type Reader interface {
    Read(p []byte) (n int, err error)
} 
```
Sometimes the result isn't correct English, but we do it anyway: 
```
type Execer interface {
    Exec(query string, args []Value) (Result, error)
}
```

Sometimes we use English to make it nicer: 
```
type ByteReader interface {
    ReadByte() (c byte, err error)
}
```

When an interface includes multiple methods, choose a name that accurately describes its purpose (examples: net.Conn, http.ResponseWriter, io.ReadWriter). 

### Errors
Error types should be of the form FooError: 
```
type ExitError struct {
    ...
}
```

Error values should be of the form ErrFoo: 
```
var ErrFormat = errors.New("image: unknown format")
```

### Packages
By convention, packages are given lower case, single-word names; there should be no need for underscores or mixedCaps.

Another convention is that the package name is the base name of its source directory; the package in src/encoding/base64 is imported as "encoding/base64" but has name base64, not encoding_base64 and not encodingBase64. 

The package name is only the default name for imports; it need not be unique across all source code, and in the rare case of a collision the importing package can choose a different name to use locally. In any case, confusion is rare because the file name in the import determines just which package is being used. 

Choose package names that lend meaning to the names they export.

Steer clear of util, common, and the like. 

### Import paths
The last component of a package path should be the same as the package name. 
```
"compress/gzip" // package gzip
```

Avoid stutter in repository and package paths: 
```
"code.google.com/p/goauth2/oauth2" // bad; my fault
```

For libraries, it often works to put the package code in the repo root:
```
"github.com/golang/oauth2" // package oauth2
```

Also avoid upper case letters (not all file systems are case sensitive). 

### Getters
Go doesn't provide automatic support for getters and setters. There's nothing wrong with providing getters and setters yourself, and it's often appropriate to do so, but it's neither idiomatic nor necessary to put Get into the getter's name. If you have a field called owner (lower case, unexported), the getter method should be called Owner (upper case, exported), not GetOwner. The use of upper-case names for export provides the hook to discriminate the field from the method. A setter function, if needed, will likely be called SetOwner. Both names read well in practice: 
```
owner := obj.Owner()
if owner != user {
    obj.SetOwner(user)
}
```

### References
* Andrew Gerrand (adg@golang.org), Google Inc. (October 2014), What's in a name? 
