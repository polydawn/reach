/*
	The maestro actor consumes a stream of {Module, Pins, WareSourcing} tuples,
	heaps them up, and executes them.

	The task tuples may also provide an Promise (simple single-result handling),
	MonitorChan (for more complex progress and log info),
	and Context (for cancellation).

	Callers tend to be either the commission system or a watcher actor.
	The commission system submits many modules in a continuing series, often
	generating more when one is done, and will eventually close the submissions;
	it will read the results from every single thing it submits and essentially
	never use per-task cancellation.
	The watcher actor(s) tend to submit one module at a time and with sizable
	delays between subsequent submissions, and may often cancel submissions
	(for example if it's already detected a new set of pins for the same module).

	The monitors are often blank when submitted by the commission or watcher
	systems.  The maestro instance can also have its own monitors supplied
	at initialization time; these will be stacked with any additional monitors
	provided by the task submission.

	Note it's also possible for a maestro to be fielding requests made by another
	maestro; this may be opaque to each of them.  For example, when there's a
	maestro on one "command-and-control" host in a distributed work pool, and
	a slaved maestro on some remote slaved host which is simply taking serialized
	task descriptions from the master and all monitoring and promises are also
	opaquely proxied back over some network channel.
	(It's also possible to do such remote resource use at the repeatr layer
	rather than the module layer, so, we'll see if this actually comes up much.)
*/
package maestro

import (
	"context"

	"github.com/warpfork/go-sup"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
)

type TaskSubmission struct {
	Context context.Context
	Promise sup.Promise
	//Monitor struct{???}

	Module       api.Module
	Pins         funcs.Pins
	WareSourcing api.WareSourcing
}
