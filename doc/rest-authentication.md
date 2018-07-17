# Glusterd2 REST API authentication

Glusterd2 REST API authentication is based on [JWT](https://jwt.io). If REST
authentication is enabled then each client request should include
`Authorization` header as example below.

    Authorization: bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJnbHVzdGVyY2xpIiwiaWF0IjoxNTMxNzI5NzQyLCJleHAiOjE1MzE3Mjk3NTJ9._WIwO7PrHUHIT62SdzfkyNjqD1GEgX2cYqN8ACZCtaw

**Note**: REST Authentication can be disabled by adding `restauth=false`
in Glusterd2 config file(Default path is
`/var/lib/glusterd2/glusterd2.toml` in case of rpm installation)

## Default user

When `glusterd2` starts for the first time, it creates the
`$GLUSTERD2_STATEDIR/auth` file which will contain the secret. If
"gluster" user group is available in the system then this `auth` file
can be read by any user in that machine who are part of "gluster"
group.

**Note**: If "gluster" user group is not available during first start
of `glusterd2` it limits the read permission to `root:root`

## glustercli

With default installation, `glustercli` will know the location of
`auth` file. `glustercli` will generate JWT token using the secret
available in `auth` file and attach it with every REST API calls.

Default installation will not require any change to use `glustercli`.

If `auth` file is in different path(When `glusterd2` is running with
custom `workdir`), then run `glustercli` by specifying
`--secret-file`. For example,

    glustercli --secret-file=/root/setup1/glusterd2/auth peer status

**Note**: bash alias can be added like `alias glustercli='glustercli
--secret-file=/root/setup1/glusterd2/auth`

Secret is taken by `glustercli` in following order of precedence
(highest to lowest):

    --secret
    --secret-file
    GLUSTERD2_AUTH_SECRET (environment variable)
    --secret-file (default path)

## Curl example

Download the utility script `glustercli-auth-header.py` from
[here](https://github.com/gluster/glusterd2/tree/master/pkg/tools/)
and save it in server nodes.

Add alias in `~/.bashrc` as below,

    alias gcurl='curl -H "$(python3 ~/glustercli-auth-header.py --secret-file=/var/lib/glusterd2/auth)"'

**Note**: Change the path of script and auth file to match your setup.

Thats all, use `gcurl` wherever `curl` is necessory. For example, to start a Volume

    gcurl -XPOST http://localhost:24007/v1/volumes/gv1/start

## Python example

Install `jwt` library using `pip install jwt` or using `dnf install
python-jwt`. Generate JWT token using,

    import jwt
    import requests

    user = "glustercli"
    secret_file = "/var/lib/glusterd2/auth"
    secret = open(secret_file).read()

    claims = {
        "iss": "glustercli",
        "iat": datetime.utcnow(),
        "exp": datetime.utcnow() + timedelta(seconds=10)
    }

    token = jwt.encode(claims, secret, algorithm='HS256')
    resp = requests.get("http://localhost:24007/v1/peers",
                        headers={"Authorization": "bearer " + token}

    peers = []
    if resp.status_code == 200:
        peers = json.loads(resp.content)


## Using REST APIs from outside the Cluster nodes

APIs for creating external users is not yet implemented. Temporarily
copy the `auth` file from one of the server nodes and place it in a
secured location.(Say `~/.glustercli/auth`)

Run `glustercli` by specifying `--secret-file`. For example,

    glustercli --secret-file=~/.glustercli/auth peer status

Or use `curl` as below

    curl -H "$(python3 ~/glustercli-auth-header.py --secret-file=~/.glustercli/auth)" \
        http://gluster1.example.com:24007/v1/peers

