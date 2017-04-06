# Coding Conventions

Please follow coding conventions and guidelines described in the following documents:

* [CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)
* [Effective Go](https://golang.org/doc/effective_go.html)

### Some more conventions

**General:**
* Keep variable names short for variables that are local to the function
* Do not export a function or variable name outside the package until you have an external consumer for it.
* Have setter or getter interfaces/methods to access/manipulate information in a different package.
* Do not use named return values in function definitions. Use only the type.

**Error Handling:**

* Use variable name `err` to denote error variable.
* Do not panic()
* Do not ignore errors using `_` variable

**Logging:**

* If a function is only invoked as part of a transaction step, always use the transaction's logger.
* The inner-most utility functions should never log. Logging must almost always be done by the caller on receiving an `error`.
* Always use log level `DEBUG` to provide useful **diagnostic information** to developers or sysadmins.
* Use log level `INFO` to provide information to users or sysadmins. This is the kind of information you'd like to log in an out-of-the-box configuration in happy scenario.
* Use log level `WARN` when something fails but there's a workaround/fallback/retry for it and/or is fully recoverable.
* Use log level `ERROR` when something occurs which is fatal to the operation, but not to the service or application.
