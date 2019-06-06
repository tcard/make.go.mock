package makegomock

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/xerrors"
)

func (g *generator) generateDescribeMethod() error {
	mockerName := g.rename + "Mocker"
	descriptorName := g.rename + "MockDescriptor"
	_, err := io.WriteString(g.w, `
// Describe lets you describe how the methods on the resulting mock are expected
// to be called and what they will return.
//
// When you're done describing methods, call Mock to get a mock that implements
// the behavior you described.
func (m *`+mockerName+`) Describe() `+descriptorName+` {
	return `+descriptorName+`{m: m}
}

// A `+descriptorName+` lets you describe how the methods on the resulting mock are expected
// to be called and what they will return.
//
// When you're done describing methods, call its Mock method to get a mock that
// implements the behavior you described.
type `+descriptorName+` struct {
	m *`+mockerName)
	if err != nil {
		return err
	}

	for _, method := range g.methods {
		methodDescName := g.rename + method.name + "MockDescriptor"
		_, err := io.WriteString(g.w, `
	descriptors_`+method.name+` []*`+methodDescName)
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(g.w, `
}

// Mock returns a mock that the `+g.name+` interface, following the behavior
// described by the descriptor methods.
//
// It also returns a function that should be called before the test is done to
// ensure that the expected number of calls to the mock methods happened. You
// can pass a *testing.T to it, since it implements the interface it wants.
func (d `+descriptorName+`) Mock() (m `+g.rename+`Mock, assert func(t interface{ Errorf(s string, args ...interface{})  }) (ok bool)) {
	assert = d.done()
	return d.m.Mock(), assert
}

func (d `+descriptorName+`) done() func(t interface{ Errorf(s string, args ...interface{})  }) bool {
	var atAssert []func() (method string, errs []string)
	type specErrs struct {
		fileLine string
		errs []string
	}
	`)
	if err != nil {
		return err
	}

	for _, method := range g.methods {
		methodDescName := g.rename + method.name + "MockDescriptor"
		methodSig := sigStr(method.sig, false)
		methodSigSpread := sigStr(method.sig, true)
		callArgNames := argNamesForCall(method.sig.args, method.sig.variadic, false)
		callArgs := strings.Join(callArgNames, ", ")

		maybeSavePrevReturns := ""
		maybeReturnPrev := ""
		maybeReturn := ""
		maybeReturnEarly := `
				return`
		if len(method.sig.ret) > 0 {
			maybeSavePrevReturns = `
			prev := desc.call`
			maybeReturnPrev = `
				return prev(` + callArgs + `)`
			maybeReturn = "return "
			maybeReturnEarly = ""
		}

		_, err = io.WriteString(g.w, `
	if len(d.descriptors_`+method.name+`) > 0 {
		for _, desc := range d.descriptors_`+method.name+` {
			desc := desc
			calls := 0`+maybeSavePrevReturns+`
			desc.call = func`+methodSig+` {
				calls++`+maybeReturnPrev+`
			}
			atAssert = append(atAssert, func() (method string, errs []string) {
				err := desc.times(calls)
				if err != nil {
					return "`+method.name+`", []string{err.Error()}
				}
				return "", nil
			})
		}
		d.m.`+method.name+` = func`+methodSigSpread+` {
			var matching []*`+methodDescName+`
			var allErrs []specErrs
			for _, desc := range d.descriptors_`+method.name+` {
				errs := desc.argValidator(`+callArgs+`)
				if len(errs) > 0 {
					allErrs = append(allErrs, specErrs{desc.fileLine, errs})
				} else {
					matching = append(matching, desc)
				}
			}
			if len(matching) == 1 {
				`+maybeReturn+`matching[0].call(`+callArgs+`)`+maybeReturnEarly+`
			}
			var args string
			for i, arg := range []interface{}{`+callArgs+`} {
				if i != 0 {
					args += "\n\t"
				}
				args += `+g.fmtPkg+`.Sprintf("%#v", arg)
			}
			if len(matching) == 0 {
				matchingErrs := ""
				for _, errs := range allErrs {
					matchingErrs += "\n\tcandidate described at "+errs.fileLine+":\n"
					for _, err := range errs.errs {
						matchingErrs += "\n\t\t" + err
					}
				}
				panic(`+g.fmtPkg+`.Errorf("no matching candidate for call to mock for `+g.rename+`.`+method.name+` with args:\n\n\t%+v\n\nfailing candidates:\n%s", args, matchingErrs))
			}
			matchingLines := ""
			for _, m := range matching {
				matchingLines += "\n\tcandidate described at " + m.fileLine
			}
			panic(`+g.fmtPkg+`.Errorf("more than one candidate for call to mock for `+g.rename+`.`+method.name+` with args:\n\n\t%+v\n\nmatching candidates:\n%s", args, matchingLines))
		}
	} else {
		d.m.`+method.name+` = func`+methodSigSpread+` {
			panic("unexpected call to mock for `+g.rename+`.`+method.name+`")
		}
	}`)
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(g.w, `
	return func(t interface{ Errorf(s string, args ...interface{})  }) bool {
		ok := true
		for _, assert := range atAssert {
			method, errs := assert()
			for _, err := range errs {
				ok = false
				t.Errorf("mock for `+g.rename+`.%s: %s", method, err)
			}
		}
		return ok
	}
}
	`)
	if err != nil {
		return err
	}

	closingMethods := [][2]string{}
	for _, method := range g.methods {
		methodDescName := g.rename + method.name + "MockDescriptor"
		closingMethods = append(closingMethods, [2]string{method.name, `*` + methodDescName})
	}

	for _, method := range g.methods {
		err := g.generateMethodDescriptor(method, closingMethods)
		if err != nil {
			return xerrors.Errorf("generating descriptor for %s: %w", g.fullName(method), err)
		}
	}

	return nil
}

func (g *generator) generateMethodDescriptor(method method, closingMethods [][2]string) error {
	descriptorName := g.rename + "MockDescriptor"
	methodDescName := g.rename + method.name + "MockDescriptor"
	argValidatorSig := validatorSig(method.sig)
	argValidatorSigStr := "func" + sigStr(argValidatorSig, false)

	_, err := io.WriteString(g.w, `
// `+method.name+` starts describing a way method `+g.rename+`.`+method.name+` is expected to be called
// and what it should return.
//
// You can call it several times to describe different behaviors, each matching different parameters.
func (d `+descriptorName+`) `+method.name+`() *`+methodDescName+` {
	return d.new`+methodDescName+`()
}

func (d `+descriptorName+`) new`+methodDescName+`() *`+methodDescName+` {
	_, file, line, _ := `+g.runtimePkg+`.Caller(2)
	return &`+methodDescName+`{
		mockDesc: d,
		times: func(int) error { return nil },
		argValidator: `+argValidatorSigStr+` { return nil },
		fileLine: `+g.fmtPkg+`.Sprintf("%s:%d", file, line),
	}
}

// `+methodDescName+` is returned by `+descriptorName+`.`+method.name+` and
// holds methods to describe the mock for method `+g.rename+`.`+method.name+`.
type `+methodDescName+` struct {
	mockDesc `+descriptorName+`
	times func(int) error
	argValidator `+argValidatorSigStr+`
	call func`+sigStr(method.sig, false)+`
	fileLine string
}
	`)
	if err != nil {
		return err
	}

	receiver := "*" + methodDescName
	methodDesc := "d"

	args := make([]argument, 0, len(method.sig.args)+1)
	for _, arg := range method.sig.args {
		args = append(args, arg)
	}
	if method.sig.variadic != nil {
		arg := *method.sig.variadic
		arg.typ = "[]" + arg.typ
		args = append(args, arg)
	}

	for i, arg := range args {
		prefix := "And"
		suffix := "Args"
		if i == 0 {
			prefix = "Takes"
			suffix = "Arg"
		}
		descriptorReturns := fmt.Sprintf("%sWith%d%s", methodDescName, i+1, suffix)

		optsArg := argument{
			name: "opts",
			typ:  g.cmpPkg + ".Option",
		}
		if arg.name == optsArg.name {
			optsArg.name += "_"
		}

		// We need to undo this randomness for the examples:
		//

		_, err := io.WriteString(g.w, `
// `+prefix+` lets you specify a value with which the actual value passed to
// the mocked method `+g.rename+`.`+method.name+` as parameter #`+fmt.Sprintf("%d", i+1)+`
// will be compared. 
//
// Package "github.com/google/go-cmp/cmp" is used to do the comparison. You can
// pass extra options for it.
//
// If you want to accept any value, use `+prefix+`Any.
//
// If you want more complex validation logic, use `+prefix+`Matching.
func (d `+receiver+`) `+prefix+`(`+argsStr([]argument{arg}, &optsArg, true)+`) `+descriptorReturns+` {
	prev := `+methodDesc+`.argValidator
	`+methodDesc+`.argValidator = `+argValidatorSigStr+` {
		errMsgs := prev(`+argsForCall(argValidatorSig.args, argValidatorSig.variadic, false)+`)
		if diff := `+g.cmpPkg+`.Diff(`+arg.name+`, got_`+arg.name+`, `+optsArg.name+`...); diff != "" {
			errMsgs = append(errMsgs, "parameter #`+fmt.Sprintf("%d", i+1)+` mismatch:\n" + diff)
		}
		return errMsgs
	}
	return `+descriptorReturns+`{`+methodDesc+`}
}

// `+prefix+`Any declares that any value passed to the mocked method
// `+method.name+` as parameter #`+fmt.Sprintf("%d", i+1)+` is expected.
func (d `+receiver+`) `+prefix+`Any() `+descriptorReturns+` {
	return `+descriptorReturns+`{`+methodDesc+`}
}

// `+prefix+`Matching lets you pass a function to accept or reject the actual
// value passed to the mocked method `+g.rename+`.`+method.name+` as parameter #`+fmt.Sprintf("%d", i+1)+`.
func (d `+receiver+`) `+prefix+`Matching(match func(`+argStr(arg, false, false)+`) error) `+descriptorReturns+` {
	prev := `+methodDesc+`.argValidator
	`+methodDesc+`.argValidator = `+argValidatorSigStr+` {
		errMsgs := prev(`+argsForCall(argValidatorSig.args, argValidatorSig.variadic, false)+`)
		if err := match(got_`+arg.name+`); err != nil {
			errMsgs = append(errMsgs, "parameter \"`+arg.name+`\" custom matcher error: " + err.Error())
		}
		return errMsgs
	}
	return `+descriptorReturns+`{`+methodDesc+`}
}

// `+descriptorReturns+` is a step forward in the description of a way that the
// method `+g.rename+`.`+method.name+` is expected to be called, with `+fmt.Sprintf("%d", i+1)+`
// arguments specified.
//
// It has methods to describe the next argument, if there's
// any left, or the return values, if there are any, or the times it's expected
// to be called otherwise.
type `+descriptorReturns+` struct {
	methodDesc *`+methodDescName+`
}
	`)
		if err != nil {
			return err
		}

		receiver = descriptorReturns
		methodDesc = "d.methodDesc"
	}

	if len(method.sig.ret) > 0 {
		descriptorReturns := methodDescName + "WithReturn"

		_, err := io.WriteString(g.w, `
// Returns lets you specify the values that the mocked method `+g.rename+`.`+method.name+`,
// if called with values matching the expectations, will return.
func (d `+receiver+`) Returns(`+argsStr(method.sig.ret, nil, false)+`) `+descriptorReturns+` {
	return d.ReturnsFrom(func`+sigStrNoNames(method.sig, false)+` {
		return `+argsForCall(method.sig.ret, nil, false)+`
	})
}

// Returns lets you specify the values that the mocked method `+g.rename+`.`+method.name+`,
// if called with values matching the expectations, will return.
// 
// It passes such passed values to a function that then returns the return values. 
func (d `+receiver+`) ReturnsFrom(f func`+sigStr(method.sig, false)+`) `+descriptorReturns+` {
	`+methodDesc+`.call = f
	return `+descriptorReturns+`{`+methodDesc+`}
}

// `+descriptorReturns+` is a step forward in the description of a way that
// method `+g.rename+`.`+method.name+` is to behave when called, with all expected parameters
// and the resulting values specified.
// arguments specified.
// 
// It has methods to describe the times the method is expected to be called,
// or you can start another method call description, or you can call Mock to
// end the description and get the resulting mock.
type `+descriptorReturns+` struct {
	methodDesc *`+methodDescName+`
}
	`)
		if err != nil {
			return err
		}

		receiver = descriptorReturns
		methodDesc = "d.methodDesc"
	}

	_, err = io.WriteString(g.w, `
// Times lets you specify a exact number of times this method is expected to be
// called.
func (d `+receiver+`) Times(times int) `+descriptorName+` {
	return d.TimesMatching(func(got int) error {
		if got != times {
			return fmt.Errorf("expected exactly %d calls, got %d", times, got)
		}
		return nil
	})
}

// AtLeastTimes lets you specify a minimum number of times this method is expected to be
// called.
func (d `+receiver+`) AtLeastTimes(times int) `+descriptorName+` {
	return d.TimesMatching(func(got int) error {
		if got < times {
			return fmt.Errorf("expected at least %d calls, got %d", times, got)
		}
		return nil
	})
}

// TimesMatching lets you pass a function to accept or reject the number of times
// this method has been called.
func (d `+receiver+`) TimesMatching(f func(times int) error) `+descriptorName+` {
	`+methodDesc+`.times = f
	`+methodDesc+`.done()
	return `+methodDesc+`.mockDesc
}

// Mock finishes the description and produces a mock.
//
// See `+g.rename+`MockDescriptor.Mock for details.
func (d `+receiver+`) Mock() (m `+g.rename+`Mock, assert func(t interface{ Errorf(string, ...interface{})  }) (ok bool)) {
	`+methodDesc+`.done()
	return `+methodDesc+`.mockDesc.Mock()
}
	`)
	if err != nil {
		return err
	}

	for _, m := range closingMethods {
		methodDescName := g.rename + m[0] + "MockDescriptor"
		_, err = io.WriteString(g.w, `
// `+m[0]+` finishes the current description for method `+g.rename+`.`+method.name+` and
// starts describing for method `+m[0]+`.
//
// See `+g.rename+`MockDescriptor.`+m[0]+` for details.
func (d `+receiver+`) `+m[0]+`() `+m[1]+` {
	`+methodDesc+`.done()
	return `+methodDesc+`.mockDesc.new`+methodDescName+`()
}
	`)
		if err != nil {
			return err
		}
	}

	_, err = io.WriteString(g.w, `
func (d *`+methodDescName+`) done() {
	d.mockDesc.descriptors_`+method.name+` = append(d.mockDesc.descriptors_`+method.name+`, d)
}
	`)
	if err != nil {
		return err
	}

	return nil
}

func (g *generator) fullName(method method) string {
	return fmt.Sprintf("%s.%s", g.name, method.name)
}

func argNamesForCall(args []argument, variadic *argument, spread bool) []string {
	names := make([]string, 0, len(args)+1)
	for _, arg := range args {
		names = append(names, arg.name)
	}
	if variadic != nil {
		name := variadic.name
		if spread {
			name += "..."
		}
		names = append(names, name)
	}
	return names
}

func argsForCall(args []argument, variadic *argument, spread bool) string {
	return strings.Join(argNamesForCall(args, variadic, spread), ", ")
}

func validatorSig(sig signature) signature {
	sig.ret = []argument{{typ: "[]string"}}
	args := make([]argument, 0, len(sig.args))
	for _, arg := range sig.args {
		arg.name = "got_" + arg.name
		args = append(args, arg)
	}
	sig.args = args
	if sig.variadic != nil {
		arg := *sig.variadic
		arg.name = "got_" + arg.name
		sig.variadic = &arg
	}
	return sig
}
