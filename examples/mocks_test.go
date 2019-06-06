package examples

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOneMatches(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	expectedErr := errors.New("expected")
	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().Takes(123).And(m).AndAny().Returns(1, expectedErr).
		ShouldBeFun().Takes(456).And(m).AndAny().Returns(2, errors.New("err 2")).
		Mock()
	defer assertMock(t)

	gotInt, gotErr := mock.ShouldBeFun(123, m)
	assert.Equal(t, 1, gotInt)
	assert.Equal(t, expectedErr, gotErr)
}

func TestNoMatches(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	expectedErr := errors.New("expected")
	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().Takes(123).And(m).AndAny().Returns(1, expectedErr).
		ShouldBeFun().Takes(456).And(m).AndAny().Returns(2, errors.New("err 2")).
		Mock()
	defer assertMock(t)

	assert.Panics(t, func() {
		mock.ShouldBeFun(789, m)
	})
}

func TestTooManyMatches(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	expectedErr := errors.New("expected")
	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().TakesAny().And(m).AndAny().Returns(1, expectedErr).
		ShouldBeFun().Takes(456).And(m).AndAny().Returns(2, errors.New("err 2")).
		Mock()
	defer assertMock(t)

	gotInt, gotErr := mock.ShouldBeFun(123, m)
	assert.Equal(t, 1, gotInt)
	assert.Equal(t, expectedErr, gotErr)

	assert.Panics(t, func() {
		mock.ShouldBeFun(456, m)
	})
}

func TestUnexpected(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().TakesAny().And(m).AndAny().Returns(1, nil).
		Mock()
	defer assertMock(t)

	assert.Panics(t, func() {
		mock.Boring()
	})
}

func TestTakesMatchingOK(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	expectedErr := errors.New("expected")
	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().
		TakesAny().
		AndMatching(func(got map[string]map[MyStruct]bool) error {
			if len(got) == 0 {
				return nil
			}
			return errors.New("bad map!")
		}).
		AndAny().
		Returns(1, expectedErr).
		Mock()
	defer assertMock(t)

	gotInt, gotErr := mock.ShouldBeFun(123, m)
	assert.Equal(t, 1, gotInt)
	assert.Equal(t, expectedErr, gotErr)
}

func TestTakesMatchingFail(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	expectedErr := errors.New("expected")
	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().
		TakesAny().
		AndMatching(func(got map[string]map[MyStruct]bool) error {
			if len(got) == 0 {
				return errors.New("bad map!")
			}
			return nil
		}).
		AndAny().
		Returns(1, expectedErr).
		Mock()
	defer assertMock(t)

	assert.Panics(t, func() {
		mock.ShouldBeFun(123, m)
	})
}

func TestReturnsFrom(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	expectedErr := errors.New("expected")

	makeRet := func(i int, m map[string]map[MyStruct]bool, _ []chan<- <-chan struct{}) (int, error) {
		return i * 2, expectedErr
	}

	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().TakesAny().And(m).AndAny().ReturnsFrom(makeRet).
		Mock()
	defer assertMock(t)

	gotInt, gotErr := mock.ShouldBeFun(2, m)
	assert.Equal(t, 4, gotInt)
	assert.Equal(t, expectedErr, gotErr)
}

func TestTimesOK(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	expectedErr := errors.New("expected")
	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().Takes(123).And(m).AndAny().Returns(1, expectedErr).Times(1).
		ShouldBeFun().Takes(456).And(m).AndAny().Returns(2, nil).Times(2).
		Mock()
	defer assertMock(t)

	gotInt, gotErr := mock.ShouldBeFun(123, m)
	assert.Equal(t, 1, gotInt)
	assert.Equal(t, expectedErr, gotErr)

	gotInt, gotErr = mock.ShouldBeFun(456, m)
	assert.Equal(t, 2, gotInt)
	assert.Nil(t, gotErr)

	gotInt, gotErr = mock.ShouldBeFun(456, m)
	assert.Equal(t, 2, gotInt)
	assert.Nil(t, gotErr)
}

func TestTimesFail(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	expectedErr := errors.New("expected")
	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().Takes(123).And(m).AndAny().Returns(1, expectedErr).Times(1).
		ShouldBeFun().Takes(456).And(m).AndAny().Returns(2, nil).Times(2).
		Mock()

	gotInt, gotErr := mock.ShouldBeFun(123, m)
	assert.Equal(t, 1, gotInt)
	assert.Equal(t, expectedErr, gotErr)

	gotInt, gotErr = mock.ShouldBeFun(456, m)
	assert.Equal(t, 2, gotInt)
	assert.Nil(t, gotErr)

	assert.Panics(t, func() {
		assertMock(fakeT(func(string, ...interface{}) {
			panic("fails!")
		}))
	})
}

func TestAtLeastTimesOK(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	expectedErr := errors.New("expected")
	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().Takes(123).And(m).AndAny().Returns(1, expectedErr).AtLeastTimes(1).
		ShouldBeFun().Takes(456).And(m).AndAny().Returns(2, nil).AtLeastTimes(1).
		Mock()
	defer assertMock(t)

	gotInt, gotErr := mock.ShouldBeFun(123, m)
	assert.Equal(t, 1, gotInt)
	assert.Equal(t, expectedErr, gotErr)

	gotInt, gotErr = mock.ShouldBeFun(456, m)
	assert.Equal(t, 2, gotInt)
	assert.Nil(t, gotErr)

	gotInt, gotErr = mock.ShouldBeFun(456, m)
	assert.Equal(t, 2, gotInt)
	assert.Nil(t, gotErr)
}

func TestAtLeastTimesFail(t *testing.T) {
	m := map[string]map[MyStruct]bool{}
	expectedErr := errors.New("expected")
	mock, assertMock := (&MyInterfaceMocker{}).Describe().
		ShouldBeFun().Takes(123).And(m).AndAny().Returns(1, expectedErr).AtLeastTimes(1).
		ShouldBeFun().Takes(456).And(m).AndAny().Returns(2, nil).AtLeastTimes(2).
		Mock()

	gotInt, gotErr := mock.ShouldBeFun(123, m)
	assert.Equal(t, 1, gotInt)
	assert.Equal(t, expectedErr, gotErr)

	gotInt, gotErr = mock.ShouldBeFun(456, m)
	assert.Equal(t, 2, gotInt)
	assert.Nil(t, gotErr)

	assert.Panics(t, func() {
		assertMock(fakeT(func(string, ...interface{}) {
			panic("fails!")
		}))
	})
}

type fakeT func(string, ...interface{})

func (f fakeT) Errorf(s string, args ...interface{}) {
	f(s, args...)
}
