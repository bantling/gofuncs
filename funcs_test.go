package gofuncs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexOf(t *testing.T) {
	// Slice
	// Index exists
	assert.Equal(t, 1, IndexOf([]int{1}, 0))
	// Index does not exist, have default
	assert.Equal(t, 2, IndexOf([]int{1}, 1, 2))
	// Index does not exist, no default
	assert.Equal(t, 0, IndexOf([]int{1}, 1))

	// Array
	// Index exists
	assert.Equal(t, 1, IndexOf([1]int{1}, 0))
	// Index does not exist, have default
	assert.Equal(t, 2, IndexOf([1]int{1}, 1, 2))
	// Index does not exist, no default
	assert.Equal(t, 0, IndexOf([1]int{1}, 1))

	func() {
		defer func() {
			assert.Equal(t, indexOfErrorMsg, recover())
		}()

		IndexOf(nil, 0)
		assert.Fail(t, "must panic")
	}()

	func() {
		defer func() {
			assert.Equal(t, indexOfErrorMsg, recover())
		}()

		IndexOf(5, 0)
		assert.Fail(t, "must panic")
	}()

	func() {
		defer func() {
			assert.Equal(t, indexOfErrorMsg, recover())
		}()

		IndexOf(5, 0)
		assert.Fail(t, "must panic")
	}()
}

func TestValueOfKey(t *testing.T) {
	// Key exists
	assert.Equal(t, 1, ValueOfKey(map[string]int{"1": 1}, "1"))
	// Key does not exist, with default
	assert.Equal(t, 2, ValueOfKey(map[string]int{"1": 1}, "", 2))
	// Key does not exist, no default
	assert.Equal(t, 0, ValueOfKey(map[string]int{"1": 1}, ""))

	func() {
		defer func() {
			assert.Equal(t, valueOfKeyErrorMsg, recover())
		}()

		ValueOfKey(nil, 0)
		assert.Fail(t, "must panic")
	}()

	func() {
		defer func() {
			assert.Equal(t, valueOfKeyErrorMsg, recover())
		}()

		ValueOfKey(5, 0)
		assert.Fail(t, "must panic")
	}()
}

func TestFilter(t *testing.T) {
	// Exact match
	filterFn := Filter(func(i interface{}) bool { return i.(int) < 3 })
	assert.True(t, filterFn(1))
	assert.False(t, filterFn(5))

	// Inexact match
	filterFn = Filter(func(i int) bool { return i < 3 })
	assert.True(t, filterFn(uint8(1)))
	assert.False(t, filterFn(5))

	// Exact and Inexact match all
	filterFns := FilterAll(
		func(i interface{}) bool { return i.(int) < 3 },
		func(i int) bool { return i >= 0 },
	)

	assert.True(t, filterFns[0](1))
	assert.False(t, filterFns[1](int8(-1)))

	// Exact and Inexact match And
	filterFn = And(
		func(i interface{}) bool { return i.(int) < 3 },
		func(i int) bool { return i >= 0 },
	)

	assert.True(t, filterFn(1))
	assert.False(t, filterFn(-1))

	// Exact and Inexact match Or
	filterFn = Or(
		func(i interface{}) bool { return i.(int) < 3 },
		func(i int) bool { return i%2 == 0 },
	)

	assert.True(t, filterFn(1))
	assert.False(t, filterFn(5))

	// Exact and Inexact match Not
	filterFn = Not(func(i interface{}) bool { return i.(int) < 3 })

	assert.False(t, filterFn(1))
	assert.True(t, filterFn(5))

	// EqualTo/DeepEqualTo
	filterFn = EqualTo(nil)
	filterFn2 := DeepEqualTo(nil)

	assert.True(t, filterFn(nil))
	assert.True(t, filterFn2(nil))
	assert.False(t, filterFn(0))
	assert.False(t, filterFn2(0))

	filterFn = EqualTo(([]int)(nil))
	filterFn2 = DeepEqualTo(([]int)(nil))

	assert.True(t, filterFn(([]int)(nil)))
	assert.True(t, filterFn2(([]int)(nil)))
	assert.False(t, filterFn(nil))
	assert.False(t, filterFn2(nil))
	assert.False(t, filterFn(([]string)(nil)))
	assert.False(t, filterFn2(([]string)(nil)))

	theVal := []int{1, 2}
	filterFn = EqualTo(theVal)
	filterFn2 = DeepEqualTo([]int{1, 2})

	assert.False(t, filterFn(([]int)(nil)))
	assert.False(t, filterFn2(([]int)(nil)))
	assert.False(t, filterFn(nil))
	assert.False(t, filterFn2(nil))
	assert.False(t, filterFn([]int{1}))
	assert.False(t, filterFn2([]int{1}))
	assert.False(t, filterFn([]int{1, 2}))
	assert.True(t, filterFn2([]int{1, 2}))
	assert.True(t, filterFn(theVal))
	assert.True(t, filterFn2(theVal))

	filterFn = EqualTo(1)
	filterFn2 = DeepEqualTo(1)

	assert.True(t, filterFn(int8(1)))
	assert.True(t, filterFn2(int8(1)))
	assert.False(t, filterFn(5))
	assert.False(t, filterFn2(5))

	// Nil
	filterFn = IsNil

	assert.True(t, filterFn(nil))
	assert.True(t, filterFn([]int(nil)))
	var f func()
	assert.True(t, filterFn(f))
	assert.False(t, filterFn(""))

	deferFunc := func() {
		assert.Equal(t, filterErrorMsg, recover())
	}

	func() {
		defer deferFunc()

		// Not a func
		Filter(0)
	}()

	func() {
		defer deferFunc()

		// Nil
		Filter(nil)
	}()

	func() {
		defer deferFunc()

		// Nil func
		var fn func()
		Filter(fn)
	}()

	func() {
		defer deferFunc()

		// No arg
		Filter(func() {})
	}()

	func() {
		defer deferFunc()

		// No result
		Filter(func(int) {})
	}()

	func() {
		defer deferFunc()

		// Wrong result type
		Filter(func(int) int { return 0 })
	}()
}

func TestMap(t *testing.T) {
	// Exact match
	mapFn := Map(func(i interface{}) interface{} { return i.(int) * 2 })
	assert.Equal(t, 2, mapFn(1))

	// Inexact match
	mapFn = Map(func(i int) int { return i * 2 })
	assert.Equal(t, 4, mapFn(uint8(2)))
	assert.Equal(t, 6, mapFn(3))

	deferFunc := func() {
		assert.Equal(t, mapErrorMsg, recover())
	}

	func() {
		defer deferFunc()

		// Not a func
		Map(0)
	}()

	func() {
		defer deferFunc()

		// Nil
		Map(nil)
	}()

	func() {
		defer deferFunc()

		// Nil func
		var fn func(int) int
		Map(fn)
	}()

	func() {
		defer deferFunc()

		// No args
		Map(func() {})
	}()

	func() {
		defer deferFunc()

		// No result
		Map(func(int) {})
	}()
}

func TestMapTo(t *testing.T) {
	// Exact match
	mapFn := MapTo(func(i interface{}) int { return i.(int) * 2 }, 0).(func(interface{}) int)
	assert.Equal(t, 2, mapFn(1))

	// Inexact match
	mapFn = MapTo(func(i int) int { return i * 2 }, 0).(func(interface{}) int)
	assert.Equal(t, 4, mapFn(2))

	// Conversion match
	mapFn = MapTo(func(i int8) int8 { return i * 2 }, 0).(func(interface{}) int)
	assert.Equal(t, 4, mapFn(2))

	// Arg of different type
	mapFn = MapTo(func(s string) int { str, _ := strconv.Atoi(s); return str }, 0).(func(interface{}) int)
	assert.Equal(t, 2, mapFn("2"))

	deferGen := func(errMsg string) func() {
		return func() {
			assert.Equal(t, errMsg, recover())
		}
	}

	func() {
		defer deferGen("val cannot be nil")()
		MapTo(nil, nil)
	}()

	func() {
		defer deferGen("val cannot be nil")()
		var p *int
		MapTo(p, p)
	}()

	// Not a function
	func() {
		defer deferGen(fmt.Sprintf(mapToErrorMsg, "int"))()
		MapTo("", 0)
	}()

	// Wrong signature
	func() {
		defer deferGen(fmt.Sprintf(mapToErrorMsg, "int"))()
		MapTo(func() {}, 0)
	}()

	// Returns uncovertible type
	func() {
		defer deferGen(fmt.Sprintf(mapToErrorMsg, "int"))()
		MapTo(func(string) string { return "" }, 0)
	}()
}

func TestConvertTo(t *testing.T) {
	convertFn := ConvertTo(int8(0))
	assert.Equal(t, int8(1), convertFn(1))
}

func TestSupplier(t *testing.T) {
	// Exact match
	supplierFn := Supplier(func() interface{} { return 2 })
	assert.Equal(t, 2, supplierFn())

	// Inexact match
	supplierFn = Supplier(func() int { return 4 })
	assert.Equal(t, 4, supplierFn())

	deferFunc := func() {
		assert.Equal(t, supplierErrorMsg, recover())
	}

	func() {
		defer deferFunc()

		// Not a func
		Supplier(0)
	}()

	func() {
		defer deferFunc()

		// Nil
		Supplier(nil)
	}()

	func() {
		defer deferFunc()

		// Nil func
		var fn func() int
		Supplier(fn)
	}()

	func() {
		defer deferFunc()

		// Has args
		Supplier(func(int) {})
	}()

	func() {
		defer deferFunc()

		// No result
		Supplier(func() {})
	}()
}

func TestSupplierOf(t *testing.T) {
	// Exact match
	supplierFn := SupplierOf(func() int { return 2 }, 0).(func() int)
	assert.Equal(t, 2, supplierFn())

	// Conversion match
	supplierFn = SupplierOf(func() int8 { return 4 }, 0).(func() int)
	assert.Equal(t, 4, supplierFn())

	deferGen := func(errMsg string) func() {
		return func() {
			assert.Equal(t, errMsg, recover())
		}
	}

	func() {
		defer deferGen("val cannot be nil")()
		SupplierOf(nil, nil)
	}()

	func() {
		defer deferGen("val cannot be nil")()
		var p *int
		SupplierOf(p, p)
	}()

	// Not a function
	func() {
		defer deferGen(fmt.Sprintf(supplierOfErrorMsg, "int"))()
		SupplierOf("", 0)
	}()

	// Wrong signature
	func() {
		defer deferGen(fmt.Sprintf(supplierOfErrorMsg, "int"))()
		SupplierOf(func() {}, 0)
	}()

	// Returns uncovertible type
	func() {
		defer deferGen(fmt.Sprintf(supplierOfErrorMsg, "int"))()
		SupplierOf(func() string { return "" }, 0)
	}()
}

func TestConsumer(t *testing.T) {
	// Exact match
	var (
		val        interface{}
		consumerFn = Consumer(func(i interface{}) { val = i })
	)
	consumerFn(2)
	assert.Equal(t, 2, val)

	// Inexact match
	consumerFn = Consumer(func(i int) { val = i })
	consumerFn(uint8(3))
	assert.Equal(t, 3, val)
	consumerFn(4)
	assert.Equal(t, 4, val)

	deferFunc := func() {
		assert.Equal(t, consumerErrorMsg, recover())
	}

	func() {
		defer deferFunc()

		// Not a func
		Consumer(0)
	}()

	func() {
		defer deferFunc()

		// Nil
		Consumer(nil)
	}()

	func() {
		defer deferFunc()

		// Nil func
		var fn func()
		Consumer(fn)
	}()

	func() {
		defer deferFunc()

		// No arg
		Consumer(func() {})
	}()

	func() {
		defer deferFunc()

		// Has result
		Consumer(func() int { return 0 })
	}()
}

func TestTernary(t *testing.T) {
	assert.Equal(t, 1, Ternary(true, 1, 2))
	assert.Equal(t, 2, Ternary(false, 1, 2))

	assert.Equal(t, 1, TernaryOf(true, func() interface{} { return 1 }, func() interface{} { return 2 }))
	assert.Equal(t, 2, TernaryOf(false, func() int { return 1 }, func() int { return 2 }))
}

func TestPanicOnError(t *testing.T) {
	var str string
	PanicOnError(json.Unmarshal([]byte(`"abc"`), &str))
	assert.Equal(t, "abc", str)

	func() {
		defer func() {
			assert.Equal(t, "unexpected end of JSON input", recover().(error).Error())
		}()

		PanicOnError(json.Unmarshal([]byte("{"), &str))
		assert.Fail(t, "json.Unmarshal must fail")
	}()

	i := PanicOnError2(strconv.Atoi("1")).(int)
	assert.Equal(t, 1, i)

	func() {
		defer func() {
			assert.Equal(t, `strconv.Atoi: parsing "a": invalid syntax`, recover().(error).Error())
		}()

		PanicOnError2(strconv.Atoi("a"))
		assert.Fail(t, "strconv must fail")
	}()
}
