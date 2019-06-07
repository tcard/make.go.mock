package examples

import "fmt"

func ExampleKeyValuesRepositoryMocker() {
	repo, assertMock := (&KeyValuesRepositoryMocker{}).Describe().
		Get().Takes("foo").Returns(42, nil).Times(1).
		Put().Takes("foo").And(43).Returns(nil).Times(1).
		Mock()
	defer assertMock(t)

	err := IncreaseCounter(repo, "foo")
	fmt.Println("Got err:", err)

	// Output:
	// Got err: <nil>
}

func IncreaseCounter(repo KeyValuesRepository, key string) error {
	value, err := repo.Get(key)
	if err != nil {
		return err
	}

	return repo.Put(key, value+1)
}

type Errorf func(string, ...interface{})

func (f Errorf) Errorf(s string, args ...interface{}) {
	f(s, args...)
}

// You can just use a *testing.T in tests.
var t Errorf = func(s string, args ...interface{}) {
	fmt.Printf(s, args...)
}
