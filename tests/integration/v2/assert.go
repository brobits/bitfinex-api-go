package tests

import (
	"reflect"
	"testing"
)

func isZeroOfUnderlyingType(x interface{}) bool {
	if x == nil {
		return true
	}
	if _, ok := x.([]string); ok {
		return true
	}
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}

func assert(t *testing.T, expected interface{}, actual interface{}) {
	prexp := reflect.ValueOf(expected)
	pract := reflect.ValueOf(actual)

	if pract.IsNil() {
		t.Errorf("nil actual value: %#v", actual)
		t.Fail()
		return
	}

	exp := prexp.Elem()
	act := pract.Elem()

	if !exp.IsValid() {
		t.Errorf("reflected expectation not valid (%#v)", expected)
		t.Fail()
	}

	if exp.Type() != act.Type() {
		t.Errorf("expected type %s, got %s", exp.Type(), act.Type())
		t.Fail()
	}

	for i := 0; i < exp.NumField(); i++ {
		expValueField := exp.Field(i)
		expTypeField := exp.Type().Field(i)

		actValueField := act.Field(i)
		actTypeField := act.Type().Field(i)

		if expTypeField.Name != actTypeField.Name {
			t.Errorf("expected type %s, got %s", expTypeField.Name, actTypeField.Name)
			t.Errorf("%#v", actual)
			t.Fail()
		}
		if isZeroOfUnderlyingType(expValueField.Interface()) {
			continue
		}
		if !isZeroOfUnderlyingType(expValueField.Interface()) && isZeroOfUnderlyingType(actValueField.Interface()) {
			t.Errorf("expected %s, but was empty", expTypeField.Name)
			t.Errorf("%#v", actual)
			t.Fail()
			return
		}
		if expValueField.Interface() != actValueField.Interface() {
			t.Errorf("expected %s %#v, got %#v", expTypeField.Name, expValueField.Interface(), actValueField.Interface())
			t.Fail()
		}
	}
	t.Logf("OK %s", exp.Type().Name())
}
