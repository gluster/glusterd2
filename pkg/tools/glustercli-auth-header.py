import sys
from argparse import ArgumentParser
from datetime import datetime, timedelta

import jwt


def jwt_token(user, secret):
    claims = dict()
    claims['iss'] = user
    claims['iat'] = datetime.utcnow()
    claims['exp'] = datetime.utcnow() + timedelta(seconds=10)

    token = jwt.encode(claims, secret, algorithm='HS256')
    return (b'Authorization: bearer ' + token).decode()


if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument("--user", default="glustercli")
    parser.add_argument("--secret-file", required=True)
    args = parser.parse_args()

    secret = ""
    try:
        with open(args.secret_file) as f:
            secret = f.read()
    except IOError as err:
        sys.stderr.write("Unable to open secret file\n")
        sys.stderr.write("Error: %s\n" % err)
        sys.exit(1)

    print(jwt_token(args.user, secret))
