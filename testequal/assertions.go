//go:build !solution

package testequal

import (
	"fmt"
)

// package main

// AssertEqual checks that expected and actual are equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true if arguments are equal.
func AssertEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if err := validateEqualArgs(expected, actual); err != nil {
		t.Errorf(fmt.Sprintf("Invalid operation: %#v == %#v (%s)",
			expected, actual, err))
		return false
	}

	if !ObjectsAreEqual(expected, actual) {
		expected, actual = formatUnequalValues(expected, actual)
		userMsg := messageFromMsgAndArgs(msgAndArgs...)

		t.Errorf("Not equal: \nexpected: %s\nactual  : %s\t%s", expected, actual, userMsg)
		return false
	}
	return true
}

// AssertNotEqual checks that expected and actual are not equal.
//
// Marks caller function as having failed but continues execution.
//
// Returns true if arguments are not equal.
func AssertNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	if err := validateEqualArgs(expected, actual); err != nil {
		return true
	}
	if ObjectsAreEqual(expected, actual) {
		expected, actual = formatUnequalValues(expected, actual)
		t.Errorf(fmt.Sprintf("Equal: \n"+
			"expected: %s\n"+
			"actual  : %s", expected, actual), msgAndArgs...)
		return false
	}
	return true
}

// RequireEqual does the same as AssertEqual but fails caller test immediately.
func RequireEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertEqual(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}

// RequireNotEqual does the same as AssertNotEqual but fails caller test immediately.
func RequireNotEqual(t T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	if !AssertNotEqual(t, expected, actual, msgAndArgs...) {
		t.FailNow()
	}
}
