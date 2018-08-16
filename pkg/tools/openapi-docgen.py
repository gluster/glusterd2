import os
from copy import copy
import sys
import yaml
import json


def parse_list(spec_dir, root_ele, lst):
    for ele in lst:
        if isinstance(ele, dict):
            parse_dict(spec_dir, root_ele, ele)


def parse_dict(spec_dir, root_ele, d):
    keys = list(d)
    for k in keys:
        if k == "$ref":
            parse_ref(spec_dir, root_ele, d)
        elif isinstance(d[k], dict):
            parse_dict(spec_dir, root_ele, d[k])
        elif isinstance(d[k], list):
            parse_list(spec_dir, root_ele, d[k])


def parse_ref(spec_dir, root_ele, ele):
    if ele["$ref"].startswith("#"):
        tree = ele["$ref"].replace("#", "").strip("/").split("/")
        d = copy(root_ele)
        for t in tree:
            d = d.get(t)

        parse_dict(spec_dir, root_ele, d)
        ele.update(d)
        del ele["$ref"]
    else:
        parts = ele["$ref"].split("#")
        filepath = parts[0].strip("../")
        data = yaml.load(open(os.path.join(spec_dir, filepath)))
        tree = parts[1].strip("/").split("/")
        d = copy(data)
        for t in tree:
            d = d.get(t)

        parse_dict(spec_dir, data, d)
        ele.update(d)
        del ele["$ref"]


def get_parameters(data):
    out = "**Parameters**\n\n"
    heading = "\n| Parameter | Description | Data Type |\n"
    heading += "| ---------- | ----------- | --------- |\n"
    content = ""
    for p in data.get("parameters", []):
        content += "| %s | *%s*. %s | %s |\n" % (
            p["name"],
            "Required" if p.get("required", False) else "Optional",
            p["description"],
            p["schema"]["type"]
        )

    if content:
        return out + heading + content + "\n"

    return out + "None\n"


def get_curl_examples(method, base_url, url, data):
    path_vars = get_in_path_vars(data)
    for k, v in path_vars.items():
        url = url.replace(k, v)

    req_bodies = get_request_body_examples(data)
    out = ""
    for b in req_bodies:
        out += b["description"] + "\n\n"
        out += "```\ncurl -X %s %s%s -d '%s'\n```\n\n" % (
            method.upper(),
            base_url,
            url,
            json.dumps(b["example"])
        )

    if out:
        return out

    return "```\ncurl -X %s %s%s\n```" % (
        method.upper(), base_url, url)


def get_request_body_examples(data):
    req = data.get("requestBody", None)

    examples = []

    if req is not None:
        # TODO: application/json is hardcoded here, script will not work
        # for other use cases
        schemas_data = req["content"]["application/json"]["schema"]
        schemas = []
        if schemas_data.get("oneOf", None) is not None:
            # Multiple request body
            for s in schemas_data["oneOf"]:
                schemas.append(s)
        else:
            schemas.append(schemas_data)

        for s in schemas:
            ex = s.get("example", {})
            if ex:
                examples.append({
                    "description": s.get("description", ""),
                    "example": ex["value"]
                })


    return examples


def get_request_body(data):
    outdata = []
    req = data.get("requestBody", None)
    out = "**Request Body**(application/json)\n\n"

    curl_examples = []

    if req is not None:
        # TODO: application/json is hardcoded here, script will not work
        # for other use cases
        schemas_data = req["content"]["application/json"]["schema"]
        schemas = []
        if schemas_data.get("oneOf", None) is not None:
            # Multiple request body
            for s in schemas_data["oneOf"]:
                schemas.append(s)
        else:
            schemas.append(schemas_data)

        expand_types = []
        expand_types_outdata = ""

        for s in schemas:
            heading = s.get("description", "")
            heading += "\n\n| Field | Description | Data Type |\n"
            heading += "| ---------- | ----------- | --------- |\n"
            content = ""
            # TODO: assumed as object since only json supported now
            for n, p in s["properties"].items():
                content += "| %s | *%s*. %s | %s |\n" % (
                    n,
                    "Required" if p.get("required", False) else "Optional",
                    p.get("description", ""),
                    p["type"],
                )

                if p["type"] == "array" and p["items"]["type"] == "object":
                    expand_types.append((n, p))
                    # Hack: Support one more level of expand
                    for k, v in p["items"]["properties"].items():
                        if v["type"] == "array" and v["items"]["type"] == "object":
                            expand_types.append((k, v))

            if content:
                outdata.append(heading + content + "\n")
                for heading_name, et in expand_types:
                    heading = heading_name
                    heading += "\n\n| Field | Description | Data Type |\n"
                    heading += "| ---------- | ----------- | --------- |\n"
                    content = ""

                    for n, p in et["items"]["properties"].items():
                        content += "| %s | *%s*. %s | %s |\n" % (
                            n,
                            "Required" if p.get("required", False) else "Optional",
                            p.get("description", ""),
                            p["type"],
                        )

                    if content:
                        expand_types_outdata += heading + content + "\n"

    if outdata:
        return out + ("\n**OR**\n".join(outdata)) + expand_types_outdata

    return out + "None\n"


def get_in_path_vars(data):
    values = {}
    for p in data.get("parameters", []):
        if p["in"] == "path":
            n = "{" + p["name"] + "}"
            ex = p.get("example", "")
            if ex:
                values[n] = "%s" % ex

    return values


def get_response_example(data):
    out = "**Sample Response**\n\n"
    for response in data.get("responses", []):
        for status_code, resp in response.items():
            out += "Status Code: %s\n\n" % status_code
            if not resp:
                continue

            ex = resp["content"]["application/json"]["schema"].get("example", {})
            if ex:
                jsondata = json.dumps(ex["value"], indent=4)
                out += "```json\n%s\n```\n" % jsondata

    return out


def main(args):
    specdir = os.path.dirname(os.path.abspath(args[1]))
    data = yaml.load(open(args[1]))
    print("# %s" % data["info"]["title"])
    print('%s' % data["info"]["description"])

    parse_dict(specdir, data, data["paths"])

    base_url = "%s/%s" % (data["servers"][0]["url"], data["info"]["version"])
    for path, d in data["paths"].items():
        for meth, meth_data in d.items():
            print('## %s' % meth_data["summary"])
            print("**URL**\n\n`%s`\n" % path)
            print("**HTTP Method**\n\n`%s`\n" % meth)
            print(get_parameters(meth_data))
            print(get_request_body(meth_data))
            print("**Sample Request**  :\n")
            print(get_curl_examples(meth, base_url, path, meth_data))
            print(get_response_example(meth_data))

if __name__ == "__main__":
    main(sys.argv)
