# Coding Conventions

Please follow coding conventions and guidelines described in the following documents:

* [CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)
* [Effective Go](https://golang.org/doc/effective_go.html)

### Some more conventions

**General:**
* Keep variable names short for variables that are local to the function.
* Do not export a function or variable name outside the package until you
  have an external consumer for it.
* Have setter or getter interfaces/methods to access/manipulate information in
  a different package.
* Do not use named return values in function definitions. Use only the type.
  Exception: defer()'d functions.

**Imports:**

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

**Error Handling:**

* Use variable name `err` to denote error variable during a function call.
* Reuse the previously declared `err` variable as long as it is in scope.
  For example, do not use `errWrite` or `errRead`.
* Do not panic().
* Do not ignore errors using `_` variable unless you know what you're doing.
* Error strings should not start with a capital letter.
* If error requires passing of extra information, you can define a new type
* Error types should end in `Error` and error variables should have `Err` as
  prefix.

**Logging:**

* If a function is only invoked as part of a transaction step, always use the
  transaction's logger.
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
