# Coding Conventions

Please follow coding conventions and guidelines described in the following documents:

* [Go proverbs](https://go-proverbs.github.io/) - highly recommended read
* [CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)
* [Effective Go](https://golang.org/doc/effective_go.html)
* [How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/)

Here's a list of some more specific conventions that are often followed in
the code and will be pointed out in the review process:

### General
* Keep variable names short for variables that are local to the function.
* Do not export a function or variable name outside the package until you
  have an external consumer for it.
* Have setter or getter interfaces/methods to access/manipulate information in
  a different package.
* Do not use named return values in function definitions. Use only the type.
  Exception: defer()'d functions.

### Imports

We use the following convention for specifying imports:

```
<import standard library packages>

<import glusterd2 packages>

<import third-party packages>
```

Example:

```go
import (
        "os"
        "path"
        "strings"
        "time"

        "github.com/gluster/glusterd2/glusterd2/daemon"
        "github.com/gluster/glusterd2/pkg/utils"
        "github.com/gluster/glusterd2/version"

        log "github.com/sirupsen/logrus"
        flag "github.com/spf13/pflag"
        config "github.com/spf13/viper"
        "golang.org/x/sys/unix"
)
```

### Error Handling

* Use variable name `err` to denote error variable during a function call.
* Reuse the previously declared `err` variable as long as it is in scope.
  For example, do not use `errWrite` or `errRead`.
* Do not panic() for errors that can be bubbled up back to user. Use panic()
  only for fatal errors which shouldn't occur.
* Do not ignore errors using `_` variable unless you know what you're doing.
* Error strings should not start with a capital letter.
* If error requires passing of extra information, you can define a new type
* Error types should end with `Error`.

### Logging

* If a function is only invoked as part of a transaction step, always use the
  transaction's logger to ensure propagation of request ID and transaction ID.
* The inner-most utility functions should never log. Logging must almost always
  be done by the caller on receiving an `error`.
* Always use log level `DEBUG` to provide useful **diagnostic information** to
  developers or sysadmins.
* Use log level `INFO` to provide information to users or sysadmins. This is the
  kind of information you'd like to log in an out-of-the-box configuration in
  happy scenario.
* Use log level `WARN` when something fails but there's a workaround or fallback
  or retry for it and/or is fully recoverable.
* Use log level `ERROR` when something occurs which is fatal to the operation,
  but not to the service or application.

### Use of goto

Use of `goto` is generally frowned up on in higher level languages. We use
`goto` statements for the following specific uses:
* Ensure RPCs always return a reply to caller
* Getting out of nested loops
* Auto-generated code

Please use `defer()` for ensuring that relevant resource cleanups happen when a
function/method exits. Also use `defer()` to revert something on a later
failure.

Developers with significant experience in C should be careful not to
excessively use `goto` just to ensure single exit point to a function. Unlike
C programs, there is no memory to be free()d here. Care must be taken when one
ports code from glusterd1 (c) to glusterd2 (go).

### glusterd2 specific conventions

**Do not log at the caller of `txn.*` methods:**

Certain patterns repeat very often in codebase. For reducing clutter, we have
moved some of the logging at the caller to inside the function. One such
instance are the methods of `transaction.Context` interface, which are used at
so many places. They log internally and caller shouldn't be logging. For
example:

```go
if err := txn.Ctx.Set("req", &req); err != nil {
        // do NOT log here
        restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err)
        return
}
```

**Use log.WithError() to log errors**

It is common pattern to log the error received as a field in the log entry.
Please use `WithError` for the same:

Do this:
```go
log.WithError(err).WithField("path", path).Error("Failed to delete path")
```

Do NOT do this:
```go
        log.WithFields(log.Fields{
                "error": err,
                "path":  path,
        }).Error("Failed to delete path")
```
