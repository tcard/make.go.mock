package examples_test

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tcard/make.go.mock/examples"
	"github.com/tcard/make.go.mock/examples/generated"
)

func Example() {
	aComplexMap := map[string]map[examples.MyStruct]bool{
		"foo": map[examples.MyStruct]bool{
			{1, nil}: false,
			{2, nil}: true,
		},
		"bar": map[examples.MyStruct]bool{
			{4, nil}: true,
		},
	}

	// You could just use MyInterfaceMocker, set up your mock methods as fields
	// there, and then call Mock to get a mock implementing MyInterface.
	//
	// However, you probably want Describe to declaratively do this for you.
	//
	// Here, we set up mocks for methods Boring and ShouldBeFun.
	mockDesc := (&generated.MyInterfaceMocker{}).Describe().
		Boring().Times(3).
		ShouldBeFun().Takes(42).And(aComplexMap).AndAny().Returns(999, errors.New("my error")).AtLeastTimes(1)

	// You don't have to put all the method descriptions in the same chain.
	mockDesc.StdSomething().
		TakesMatching(func(f *os.File) error {
			if f.Name() != "expected" {
				return errors.New("unexpected!")
			}
			return nil
		}).
		AndAny().
		Returns(true).
		AtLeastTimes(1)

	// Finall, call Mock to get your mock, and pass your *testing.T to the
	// function it returns.defer
	mock, assertMock := mockDesc.Mock()
	defer assertMock(t)

	// The mock implements the interface it's mocking. Duh!
	var _ examples.MyInterface = mock

	// OK, now you would pass the mock to the real code that you're testing.
	// For this example, let's use the mock ourselves.

	// First, call Boring more times than expected. This will fail later.
	for i := 0; i < 5; i++ {
		mock.Boring()
	}

	// This call to ShouldBeFun matches one of the descriptions above, so its
	// specified return values will be returned here.
	fmt.Print("Calling ShouldBeFun with expected parameters: ")
	fmt.Println(mock.ShouldBeFun(42, aComplexMap, make(chan<- <-chan struct{}), nil))
	fmt.Println()

	// This call to ShouldBeFun doesn't match any description, so we don't know
	// what to return. It fails too.
	func() {
		fmt.Println("Calling ShouldBeFun with unexpected parameters:")

		defer func() { fmt.Println(firstLine(recover())) }()
		mock.ShouldBeFun(12345, nil)
	}()
	fmt.Println()

	// Calling a method you haven't described also fails.
	func() {
		fmt.Println("Calling ReturnSomethingAtLeast, which is undescribed:")

		defer func() { fmt.Println(firstLine(recover())) }()
		mock.ReturnSomethingAtLeast()
	}()
	fmt.Println()

	fmt.Println("Finished! Now we'll get a deferred error for having called Boring too much.")

	// Output:
	// Calling ShouldBeFun with expected parameters: 999 my error
	//
	// Calling ShouldBeFun with unexpected parameters:
	// no matching candidate for call to mock for MyInterface.ShouldBeFun with args:
	//
	// Calling ReturnSomethingAtLeast, which is undescribed:
	// unexpected call to mock for MyInterface.ReturnSomethingAtLeast
	//
	// Finished! Now we'll get a deferred error for having called Boring too much.
	// mock for MyInterface.Boring: expected exactly 3 calls, got 5
}

type Errorf func(string, ...interface{})

func (f Errorf) Errorf(s string, args ...interface{}) {
	f(s, args...)
}

// You can just use a *testing.T in tests.
var t Errorf = func(s string, args ...interface{}) {
	fmt.Printf(s, args...)
}

func firstLine(s interface{}) string {
	return strings.SplitN(fmt.Sprint(s), "\n", 2)[0]
}
