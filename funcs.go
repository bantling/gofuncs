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

	return func(arg interface{}) bool {
		var (
			argVal = reflect.ValueOf(arg)
			resVal = vfn.Call([]reflect.Value{argVal})[0].Bool()
		)

		return resVal
	}
}

// Map (fn) adapts a func(any) any into a func(interface{}) interface{}.
// If fn happens to be a func(interface{}) interface{}, it is returned as is.
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

	return func(arg interface{}) interface{} {
		var (
			argVal = reflect.ValueOf(arg)
			resVal = vfn.Call([]reflect.Value{argVal})[0].Interface()
		)

		return resVal
	}
}

// MapTo (fn, X) adapts a func(any) X' into a func(interface{}) X.
// If fn happens to be a func(interface{}) X, it is returned as is.
// Otherwise, type X' must be convertible to X.
// The result will have to be type asserted by the caller.
func MapTo(fn interface{}, val interface{}) interface{} {
	// val cannot be nil
	if (val == nil) || (fmt.Sprintf("%p", val) == "0x0") {
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
	if (val == nil) || (fmt.Sprintf("%p", val) == "0x0") {
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

	return func(arg interface{}) {
		argVal := reflect.ValueOf(arg)
		vfn.Call([]reflect.Value{argVal})
	}
}
