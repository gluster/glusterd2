/*
Package commands implements rest end points for each commands.

GlusterD 2.0 uses [mux](github.com/gorilla/mux) which implements a request
router and dispatcher. The name mux stands for "HTTP request multiplexer". Like
the standard http.ServeMux, mux.Router matches incoming requests against a list
of registered routes and calls a handler for the route that matches the URL or
other conditions.

Route models a route to be set on the GlusterD Rest server and holds the name,
pattern and the registered handler function. Individual commands should define
its route table and glusterd while initializing will iterate over all these
router tables and register them. The handler function is the one which holds the
logic of how the command is going to be executed.

Developers are expected to create individual directories for every commands
under command folder in the code repository and put up a .go file with the same
command name. Along with that individual commands must have an entry in the
Commands interface table in command.go.
*/
package commands
