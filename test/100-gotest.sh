#!/bin/bash

GOPACKAGES="$(go list ./... | grep -v vendor | grep -v e2e)"
# no special options, exec to go test w/ all pkgs
if [[ ${GD2_TEST_EXITFIRST} != "yes" && -z ${GD2_TEST_COVERAGE} ]]; then
	# shellcheck disable=SC2086
	exec go test ${GOPACKAGES}
fi

# our options are set so we need to handle each go package one
# at at time
if [[ ${GD2_TEST_COVERAGE} ]]; then
	GOTESTOPTS="-covermode=count -coverprofile=cover.out"
	COVERFILE=packagecover.out
	echo "mode: count" > "${COVERFILE}"
fi

failed=0
for gopackage in ${GOPACKAGES}; do
	echo "--- testing: ${gopackage} ---"
	# shellcheck disable=SC2086
	go test ${GOTESTOPTS} "${gopackage}"
	[ $? -ne 0 ] && ((failed+=1))
	if [[ -f cover.out ]]; then
		# Append to coverfile
		grep -v "^mode: count" cover.out >> "${COVERFILE}"
	fi
	if [[ ${GD2_TEST_COVERAGE} = "stdout" && -f cover.out ]]; then
		go tool cover -func=cover.out
	fi
	if [[ ${GD2_TEST_COVERAGE} = "html" && -f cover.out ]]; then
		mkdir -p coverage
		fn="coverage/${gopackage////-}.html"
		echo " * generating coverage html: ${fn}"
		go tool cover -html=cover.out -o "${fn}"
	fi
	rm -f cover.out
	if [[ ${failed} -ne 0 && ${GD2_TEST_EXITFIRST} = "yes" ]]; then
		exit ${failed}
	fi
done
exit ${failed}
