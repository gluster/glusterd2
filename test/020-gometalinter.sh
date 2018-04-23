#!/bin/bash

gometalinter -D gotype -E gofmt --errors --deadline=5m -j 4 --vendor
