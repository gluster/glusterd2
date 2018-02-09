#!/usr/bin/python
"""
A tool to bootstrap gd2 plugin.

A plugin can add REST routes, sunrpc methods and register transactions. This
tool bootstraps the plugin code to start with.

* Usage

    ./new-gd2-plugin.py <plugin-dir> <plugin-name>

This will create a directory for plugin and creates two files as below

    $PLUGINS_DIR/
        - <plugin-name>/
            - init.go
              rest.go

Also initializes new plugin in `$PLUGINS_DIR/plugins.go` file.

To introduce new REST API, register new REST route in `init.go` file and add
implementation in `rest.go` file. Refer the sample route added as part of
bootstrap.
"""

import sys
import string
import os
import argparse

src_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
PLUGINS_FILE = os.path.join(src_dir, "glusterd2/plugin/plugins.go")
PLUGINS_DIR = os.path.join(src_dir, "plugins")
INIT_FILE = "init.go"
REST_FILE = "rest.go"

IMPORT_PFX = "\t\"github.com/gluster/glusterd2/plugins/"

INIT_GO_TMPL = """package $name

import (
	"github.com/gluster/glusterd2/glusterd2/servers/rest/route"
	"github.com/gluster/glusterd2/pkg/sunrpc"
)

// Plugin is a structure which implements GlusterdPlugin interface
type Plugin struct {
}

// Name returns name of plugin
func (p *Plugin) Name() string {
	return "$name"
}

// SunRPCProgram returns sunrpc program to register with Glusterd
func (p *Plugin) SunRPCProgram() sunrpc.Program {
	return nil
}

// RestRoutes returns list of REST API routes to register with Glusterd
func (p *Plugin) RestRoutes() route.Routes {
	return route.Routes{
		route.Route{
			Name:        "${namec}Help",
			Method:      "GET",
			Pattern:     "/${name}/help",
			Version:     1,
			HandlerFunc: ${name}HelpHandler},
	}
}

// RegisterStepFuncs registers transaction step functions with
// Glusterd Transaction framework
func (p *Plugin) RegisterStepFuncs() {
	return
}
"""

REST_GO_TMPL = """package ${name}

import (
	"net/http"

	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
)

func ${name}HelpHandler(w http.ResponseWriter, r *http.Request) {
	// Implement the help logic and send response back as below
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, "${namec} Help")
}
"""


def generate_init_go_file(name):
    data = string.Template(INIT_GO_TMPL).substitute({
        "name": name,
        "namec": name[0].upper() + name[1:]
    })

    with open(os.path.join(PLUGINS_DIR, name, INIT_FILE), "w") as f:
        f.write(data)


def generate_rest_go_file(name):
    data = string.Template(REST_GO_TMPL).substitute({
        "name": name,
        "namec": name[0].upper() + name[1:]
    })

    with open(os.path.join(PLUGINS_DIR, name, REST_FILE), "w") as f:
        f.write(data)


def add_to_plugins_go(name):
    import_path = IMPORT_PFX + name + "\""
    add_plugin = "\t&" + name + ".Plugin{},"
    data = []
    import_started = False
    add_plugin_started = False

    with open(PLUGINS_FILE) as f:
        for line in f:
            if import_started:
                if line.strip().endswith(name + '"'):
                    print "Plugin with name \"%s\" already exists" % name
                    sys.exit(1)

                if line.strip() == ")":
                    data.append(import_path)
                    import_started = False

            if add_plugin_started:
                if line.strip() == "}":
                    data.append(add_plugin)

            if not import_started and line.strip() == "import (":
                import_started = True

            if not add_plugin_started and \
               line.strip() == "var PluginsList = []GlusterdPlugin{":
                    add_plugin_started = True

            data.append(line.strip("\n"))

    with open(PLUGINS_FILE + ".tmp", "w") as f:
        f.write("\n".join(data))


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description=__doc__)
    parser.add_argument("plugin_name", help="Plugin Name")
    args = parser.parse_args()
    add_to_plugins_go(args.plugin_name)
    try:
        os.mkdir(os.path.join(PLUGINS_DIR, args.plugin_name))
    except OSError as e:
        print "Unable to create plugin dir \"%s/%s\": %s" % (PLUGINS_DIR,
                                                             args.plugin_name,
                                                             e)

    generate_init_go_file(args.plugin_name)
    generate_rest_go_file(args.plugin_name)
    os.rename(PLUGINS_FILE + ".tmp", PLUGINS_FILE)
