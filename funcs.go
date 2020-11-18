package gofuncs

import (
	"fmt"
	"reflect"
)

const (
	filterErrorMsg     = "fn must be a non-nil function of one argument of any type that returns bool"
	mapErrorMsg        = "fn must be a non-nil function of one argument of any type that returns one value of any type"
	mapToErrorMsg      = "fn must be a non-nil function of one argument of any type that returns one value convertible to type %s"
	supplierErrorMsg   = "fn must be a non-nil function of no arguments that returns one value of any type"
	supplierOfErrorMsg = "fn must be a non-nil function of no arguments that returns one value convertible to type %s"
	consumerErrorMsg   = "fn must be a non-nil funciton of one argument of any type and no return values"
)

// Filter (fn) adapts a func(any) bool into a func(interface{}) bool.
// If fn happens to be a func(interface{}) bool, it is returned as is.
// Otherwise, each invocation converts the arg passed to the type the func receives.
func Filter(fn interface{}) func(interface{}) bool {
	// Return fn as is if it is desired type
	if res, isa := fn.(func(interface{}) bool); isa {
		return res
	}

	vfn := reflect.ValueOf(fn)
	if (vfn.Kind() != reflect.Func) || vfn.IsNil() {
		panic(filterErrorMsg)
	}

	typ := vfn.Type()
	if (typ.NumIn() != 1) ||
		(typ.NumOut() != 1) ||
		(typ.Out(0).Kind() != reflect.Bool) {
		panic(filterErrorMsg)
	}

	argTyp := typ.In(0)

	return func(arg interface{}) bool {
		var (
			argVal = reflect.ValueOf(arg).Convert(argTyp)
			resVal = vfn.Call([]reflect.Value{argVal})[0].Bool()
		)

		return resVal
	}
}

// FilterAll (fns) adapts any number of func(any) bool into a slice of func(interface{}) bool.
// Each func passed is separately adapted using Filter into the corresponding slice element of the result.
// FIlterAll is the basis for composing multiple logic functions into a single logic function.
// Note that when calling the provided set of logic functions, the argument type must be compatible with all of them.
// The most likely failure case is mixing funcs that accept interface{} that type assert the argument with funcs that accept a specific type.
func FilterAll(fns ...interface{}) []func(interface{}) bool {
	// Create adapters
	adaptedFns := make([]func(interface{}) bool, len(fns))
	for i, fn := range fns {
		adaptedFns[i] = Filter(fn)
	}

	return adaptedFns
}

// And (fns) any number of func(any)bool into the conjunction of all the funcs.
// Short-circuit logic will return false on the first function that returns false.
func And(fns ...interface{}) func(interface{}) bool {
	adaptedFns := FilterAll(fns...)

	return func(val interface{}) bool {
		for _, fn := range adaptedFns {
			if !fn(val) {
				return false
			}
		}

		return true
	}
}

// Or (fns) any number of func(any)bool into the disjunction of all the funcs.
// Short-circuit logic will return true on the first function that returns true.
func Or(fns ...interface{}) func(interface{}) bool {
	adaptedFns := FilterAll(fns...)

	return func(val interface{}) bool {
		for _, fn := range adaptedFns {
			if fn(val) {
				return true
			}
		}

		return false
	}
}

// Not (fn) adapts a func(any) bool to the negation of the func.
func Not(fn interface{}) func(interface{}) bool {
	adaptedFn := Filter(fn)

	return func(val interface{}) bool {
		return !adaptedFn(val)
	}
}

// EqualTo (val) returns a func(interface{}) bool that returns true if the func arg is equal to val.
// The arg is converted to the type of val first, then compared.
// Panics is val is nil.
func EqualTo(val interface{}) func(interface{}) bool {
	if IsNil(val) {
		panic("val cannot be nil")
	}

	typ := reflect.TypeOf(val)

	return func(arg interface{}) bool {
		return (!IsNil(arg)) && reflect.ValueOf(arg).Convert(typ).Interface() == val
	}
}

// IsNil is a func(interface{}) bool that returns true is val is nil
func IsNil(val interface{}) bool {
	if val == nil {
		return true
	}

	// Sometimes a nil value received as an empty interface doesn't compare to nil with ==, but the pointer address will be the string 0x0.
	// If the value is a string, it will print in pointer format as "%!p(string=X)", where X is the string value.
	return fmt.Sprintf("%p", val) == "0x0"
}

// Map (fn) adapts a func(any) any into a func(interface{}) interface{}.
// If fn happens to be a func(interface{}) interface{}, it is returned as is.
// Otherwise, each invocation converts the arg passed to the type the func receives.
func Map(fn interface{}) func(interface{}) interface{} {
	// Return fn as is if it is desired type
	if res, isa := fn.(func(interface{}) interface{}); isa {
		return res
	}

	vfn := reflect.ValueOf(fn)
	if (vfn.Kind() != reflect.Func) || vfn.IsNil() {
		panic(mapErrorMsg)
	}

	typ := vfn.Type()
	if (typ.NumIn() != 1) || (typ.NumOut() != 1) {
		panic(mapErrorMsg)
	}

	argTyp := typ.In(0)

	return func(arg interface{}) interface{} {
		var (
			argVal = reflect.ValueOf(arg).Convert(argTyp)
			resVal = vfn.Call([]reflect.Value{argVal})[0].Interface()
		)

		return resVal
	}
}

// MapTo (fn, X) adapts a func(any) X' into a func(interface{}) X.
// If fn happens to be a func(interface{}) X, it is returned as is.
// Otherwise, each invocation converts the arg passed to the type the func receives, and type X' must be convertible to X.
// The result will have to be type asserted by the caller.
func MapTo(fn interface{}, val interface{}) interface{} {
	// val cannot be nil
	if IsNil(val) {
		panic("val cannot be nil")
	}

	// Verify val is a non-interface type
	var (
		xval = reflect.ValueOf(val)
		xtyp = xval.Type()
	)
	if xval.Kind() == reflect.Interface {
		panic("val cannot be an interface{} value")
	}

	// Verify fn has is a non-nil func of 1 parameter and 1 result
	var (
		vfn    = reflect.ValueOf(fn)
		errMsg = fmt.Sprintf(mapToErrorMsg, xtyp)
	)

	if (vfn.Kind() != reflect.Func) || vfn.IsNil() {
		panic(errMsg)
	}

	// The func has to accept 1 arg and return 1 type
	typ := vfn.Type()
	if (typ.NumIn() != 1) || (typ.NumOut() != 1) {
		panic(errMsg)
	}

	var (
		argTyp = typ.In(0)
		resTyp = typ.Out(0)
	)

	// Return fn as is if it is desired type
	if (argTyp.Kind() == reflect.Interface) && (resTyp == xtyp) {
		return fn
	}

	// If fn returns any type convertible to X, then generate a function of interface{} to exactly X
	if !resTyp.ConvertibleTo(xtyp) {
		panic(errMsg)
	}

	return reflect.MakeFunc(
		reflect.FuncOf(
			[]reflect.Type{reflect.TypeOf((*interface{})(nil)).Elem()},
			[]reflect.Type{xtyp},
			false,
		),
		func(args []reflect.Value) []reflect.Value {
			var (
				argVal = reflect.ValueOf(args[0].Interface()).Convert(argTyp)
				resVal = vfn.Call([]reflect.Value{argVal})[0].Convert(xtyp)
			)

			return []reflect.Value{resVal}
		},
	).Interface()
}

// Supplier (fn) adapts a func() any into a func() interface{}.
// If fn happens to be a func() interface{}, it is returned as is.
func Supplier(fn interface{}) func() interface{} {
	// Return fn as is if it is desired type
	if res, isa := fn.(func() interface{}); isa {
		return res
	}

	// Verify fn has is a non-nil func of 0 parameters and 1 result
	vfn := reflect.ValueOf(fn)

	if (vfn.Kind() != reflect.Func) || vfn.IsNil() {
		panic(supplierErrorMsg)
	}

	// The func has to accept no args and return 1 type
	typ := vfn.Type()
	if (typ.NumIn() != 0) || (typ.NumOut() != 1) {
		panic(supplierErrorMsg)
	}

	return func() interface{} {
		resVal := vfn.Call([]reflect.Value{})[0].Interface()

		return resVal
	}
}

// SupplierOf (fn, X) adapts a func() X' into a func() X.
// If fn happens to be a func() X, it is returned as is.
// Otherwise, type X' must be convertible to X.
// The result will have to be type asserted by the caller.
func SupplierOf(fn interface{}, val interface{}) interface{} {
	// val cannot be nil
	if IsNil(val) {
		panic("val cannot be nil")
	}

	// Verify val is a non-interface type
	var (
		xval = reflect.ValueOf(val)
		xtyp = xval.Type()
	)
	if xval.Kind() == reflect.Interface {
		panic("val cannot be an interface{} value")
	}

	// Verify fn has is a non-nil func of 0 parameters and 1 result
	var (
		vfn    = reflect.ValueOf(fn)
		errMsg = fmt.Sprintf(supplierOfErrorMsg, xtyp)
	)

	if (vfn.Kind() != reflect.Func) || vfn.IsNil() {
		panic(errMsg)
	}

	// The func has to accept no args and return 1 type
	typ := vfn.Type()
	if (typ.NumIn() != 0) || (typ.NumOut() != 1) {
		panic(errMsg)
	}

	resTyp := typ.Out(0)

	// Return fn as is if it is desired type
	if resTyp == xtyp {
		return fn
	}

	// If fn returns any type convertible to X, then generate a function that returns exactly X
	if !resTyp.ConvertibleTo(xtyp) {
		panic(errMsg)
	}

	return reflect.MakeFunc(
		reflect.FuncOf(
			[]reflect.Type{},
			[]reflect.Type{xtyp},
			false,
		),
		func(args []reflect.Value) []reflect.Value {
			resVal := vfn.Call([]reflect.Value{})[0].Convert(xtyp)

			return []reflect.Value{resVal}
		},
	).Interface()
}

// Consumer (fn) adapts a func(any) into a func(interface{})
// If fn happens to be a func(interface{}), it is returned as is.
// Otherwise, each invocation converts the arg passed to the type the func receives.
func Consumer(fn interface{}) func(interface{}) {
	// Return fn as is if it is desired type
	if res, isa := fn.(func(interface{})); isa {
		return res
	}

	// Verify fn has is a non-nil func of 1 parameters and no result
	vfn := reflect.ValueOf(fn)

	if (vfn.Kind() != reflect.Func) || vfn.IsNil() {
		panic(consumerErrorMsg)
	}

	// The func has to accept one arg and return nothing
	typ := vfn.Type()
	if (typ.NumIn() != 1) || (typ.NumOut() != 0) {
		panic(consumerErrorMsg)
	}

	argTyp := typ.In(0)

	return func(arg interface{}) {
		argVal := reflect.ValueOf(arg).Convert(argTyp)
		vfn.Call([]reflect.Value{argVal})
	}
}
