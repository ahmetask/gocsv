# Supported Field Types

```go
type AnyStruct struct {
	F1 *string
	//...
}
type MyStruct struct {
	F1 []string
	F2 int
	F3 *int
	F4 int64
	F5 *int64
	F6 uint64
	F7 *uint64
	F8 string
	F9 *string
	F10 AnyStruct
	F11 *AnyStruct
	F12 *bool
	F13 bool
	F14 *float64
	F16 float64
}
```