#!/bin/bash

# This "test" exists merely to exercise the test "framework"

if [ "$GD2_TEST_TEST" ]; then
	exit "$GD2_TEST_TEST"
fi
exit 0
