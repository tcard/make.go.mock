# make.go.mock [![Build Status](https://secure.travis-ci.org/tcard/make.go.mock.svg?branch=master)](http://travis-ci.org/tcard/make.go.mock) [![GoDoc](https://godoc.org/github.com/tcard/make.go.mock?status.svg)](https://godoc.org/github.com/tcard/make.go.mock) [![codecov](https://codecov.io/gh/tcard/make.go.mock/branch/master/graph/badge.svg?token=)](https://codecov.io/gh/tcard/make.go.mock)

Command make.go.mock generates type-safe mocks for Go interfaces and functions.

## Installing

```
go get github.com/tcard/make.go.mock
```

## Usage

make.go.mock takes a type name (interface or function) as argument and generates Go code to mock it.

It's intended to be used with `go generate`.

For a full list of flags:

```
make.go.mock -h
```

See [examples/examples.go](https://github.com/tcard/make.go.mock/tree/master/examples/examples.go) for actual examples of `go:generate` directives.

Check out also [a full example in the docs](https://godoc.org/github.com/tcard/make.go.mock/examples#ex-package), or [the generated API for the examples package](https://godoc.org/github.com/tcard/make.go.mock/examples/generated).

## Failure feedback

make.go.mock leverages [github.com/google/go-cmp](https://github.com/google/go-cmp) to provide rich error feedback on unexpected values. It also tries to provide enough context information to help track down mistakes easily.

For instance, look at this code:

```go
m := map[string]map[MyStruct]bool{}
expectedErr := errors.New("expected")
mock, assertMock := (&MyInterfaceMocker{}).Describe().
	ShouldBeFun().Takes(123).And(m).AndAny().Returns(1, expectedErr).
	ShouldBeFun().Takes(456).And(m).AndAny().Returns(2, errors.New("err 2")).
	Mock()
defer assertMock(t)

_, _ = mock.ShouldBeFun(789, m)
```

The call `mock.ShouldBeFun(789, m)` doesn't match either of the described expectations, so the mock doesn't know what to return. It then panics like this:

```
panic: no matching candidate for call to mock for MyInterface.ShouldBeFun with args:

	789
	map[string]map[examples.MyStruct]bool{}
	[]chan<- <-chan struct {}(nil)

failing candidates:

	candidate described at /path/to/make.go.mock/examples/mocks_test.go:28:

		parameter #1 mismatch:
  int(
- 	123,
+ 	789,
  )

	candidate described at /path/to/make.go.mock/examples/mocks_test.go:29:

		parameter #1 mismatch:
  int(
- 	456,
+ 	789,
  )
```

## Comparison with other Go mocking generators

### github.com/golang/mock/mockgen

For this method:

```go
type Index interface {
	Put(key string, value interface{}) (bool, error)
}
```

mockgen generates this:

```go
func (mr *MockIndexMockRecorder) Put(arg0, arg1 interface{}) *gomock.Call
```

The string arguments are turned into `interface{}`, **losing type safety** in the process. So if you change the type of the arguments, the test code will still compile, only to crash when run. It's even more fragile if the return types or number of arguments change.

Compare this with the code generated by make.go.mock:

```go
func (d IndexMockDescriptor) GetTwo() *IndexGetTwoMockDescriptor

func (d *IndexPutMockDescriptor) Takes(key string, opts ...cmp.Option) IndexPutMockDescriptorWith1Arg
func (d *IndexPutMockDescriptor) TakesAny() IndexPutMockDescriptorWith1Arg
func (d *IndexPutMockDescriptor) TakesMatching(match func(key string) error) IndexPutMockDescriptorWith1Arg

func (d IndexPutMockDescriptorWith1Arg) And(value interface{}, opts ...cmp.Option) IndexPutMockDescriptorWith2Args
func (d IndexPutMockDescriptorWith1Arg) AndAny() IndexPutMockDescriptorWith2Args
func (d IndexPutMockDescriptorWith1Arg) AndMatching(match func(value interface{}) error) IndexPutMockDescriptorWith2Args
```

The original types are kept, and type information is preserved.

mockgen **doesn't support mocking function types**, while make.go.mock does.

### github.com/vektra/mockery

mockery generates mocks that leverage github.com/stretchr/testify/mock, which isn't a mock generator, but just a runtime library.

mockery only does part of the job that a mock generator could do, though: it doesn't generate the code to describe what the mock does. For that, you're supposed to use plain testify/mock.

This means that **type-safety is completely lost**. To describe mocks, you use methods like these:

```go
func (m *Mock) On(methodName string, arguments ...interface{}) *Call
func (c *Call) Return(returnArguments ...interface{}) *Call
```

If the method name, number and/or type of arguments and/or return values change, you'll only notice your mocks are broken when you run the tests.

mockery also **doesn't support mocking function types**, while make.go.mock does.

### Use case comparison

Now let's see how you would use each:

**mockgen**

```go
ctrl := gomock.NewController(t)
defer ctrl.Finish()

mockIndex := mock_user.NewMockIndex(ctrl)
mockIndex.EXPECT().Put("a", 1).Return(true, nil)
mockIndex.EXPECT().Put("b", gomock.Any()).Return(false, errors.New("broken"))

// Use mock
_, _ = mockIndex.Put("a", 1)
```

**mockery**

```go
mockIndex := &MockIndex{}
defer mockIndex.AssertExpectations()
mockIndex.On("Put", "a", 1).Returns(true, nil)
mockIndex.On("Put", "b", mock.Anything).Returns(false, errors.New("broken"))

// Use mock
_, _ = mockIndex.Put("a", 1)
```

**make.go.mock**

```go
mockIndex, assertMock := (&IndexMocker{}).Describe().
	Put().Takes("a").And(1).Returns(true, nil).AtLeastTimes(1).
	Put().Takes("b").AndAny().Returns(false, errors.New("broken")).
	Mock()
defer assertMock(t)

// Use mock
_, _ = mockIndex.Put("a", 1)
```

make.go.mock is more verbose; it's the tradeoff it makes to keep type safety by generating several methods per argument. Otherwise, by cramming all use cases in a single method, you force the method to take `interface{}`s in order to express things like "any value".
