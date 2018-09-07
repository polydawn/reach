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
	"github.com/warpfork/go-sup"

	"go.polydawn.net/go-timeless-api"
	"go.polydawn.net/go-timeless-api/funcs"
	"go.polydawn.net/go-timeless-api/repeatr/client/exec"
	"go.polydawn.net/stellar/gadgets/module"
)

type TaskSubmission struct {
	CancelChan <-chan struct{}
	Promise    sup.Promise
	//Monitor struct{???}

	Module       api.Module
	Pins         funcs.Pins
	WareSourcing api.WareSourcing
}

func New(Inbox <-chan TaskSubmission, StagingWarehouse api.WarehouseLocation) sup.Supervisor {
	return maestro{nil, Inbox, StagingWarehouse}.init()
}

type maestro struct {
	sup.Supervisor // Maestro *is* a task itself, composed of internal subtasks. // it's recommended to make anything that embeds a supervisor like this into an unexported type, and return only non-concrete references to it; this prevents code external to the package from casting to a concrete type and being able to mutate this field (which would be insane).

	// -- wiring --

	Inbox <-chan TaskSubmission

	// -- config --

	StagingWarehouse api.WarehouseLocation
}

func (m maestro) init() *maestro {
	millFeed := make(chan sup.Task)
	m.Supervisor = sup.SuperviseForkJoin("maestro", []sup.Task{
		&maestroControl{&m, millFeed},
		sup.SuperviseStream("mill", millFeed),
	})
	return &m
}

type maestroControl struct {
	*maestro
	millFeed chan<- sup.Task
}

func (mc *maestroControl) Name() string { return "ctrl" }
func (mc *maestroControl) Run(ctx sup.Context) error {
	defer close(mc.millFeed) // so when you have a millfeedcontrol actor, this is common: you don't want to forget to close on your way out or your sibling task that's the mill supervisor will hang.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-mc.Inbox:
			if !ok {
				return nil
			}
			mc.millFeed <- maestroTask{
				msg,
				api.WareStaging{ByPackType: map[api.PackType]api.WarehouseLocation{"tar": mc.StagingWarehouse}},
			}

			// alternatively:
			//
			//    fn := maestroTask{msg, ...}.Run
			//    go mc.millSupervisor.NewTask("%").Bind(fn).Do()
			//
			// i'm disconcerted by channels here because if this had backpressure of any kind, we'd be in trouble doing this without a ctxdone-select.  whereas a mgr.newTask design could either panic (like write to close channel already would)
			// and it's not just the ctxdone-select being syntactically annoying: if i had too many tasks to submit more, i want to return and groom my queue.  though to be fair, i guess i can do that with a nonblocking select, and that actually *is* something that's application-domain-specific maestro work.
			//
			// the *nice* thing about channels as a task submission design is that it immediately makes it clear to people that there must be a single responsible individual for closing it, and you cannot multiclose it, and nobody can argue about it.
		}
	}
}

type maestroTask struct {
	TaskSubmission
	wareStaging api.WareStaging
}

func (mt maestroTask) Run(ctx sup.Context) (_ error) {
	// if it was already resolved, that quietly takes precident, so this is safe.
	// however, this is also the kind of thing you'd want to be able to do if your task got rejected entirely.
	defer mt.Promise.Cancel()
	// Either the supervisor can cancel our work, or the original submitter of
	//  the task might say it no long cares.  Either causes us to abort.
	select {
	case <-ctx.Done():
		return
	case <-mt.CancelChan:
		return
	default:
	}
	// FIXME closure of mt.CancelChan should also work *later*, so we need a composed ctx here and to park a goroutine on forwarding it :I

	ord, err := funcs.ModuleOrderStepsDeep(mt.TaskSubmission.Module)
	if err != nil {
		mt.Promise.Resolve(err) // TODO should be a union type for this
		return
	}
	exports, err := module.Evaluate(
		ctx,
		mt.TaskSubmission.Module,
		ord,
		mt.TaskSubmission.Pins,
		mt.TaskSubmission.WareSourcing,
		mt.wareStaging,
		repeatrclient.Run,
	)
	if err != nil {
		mt.Promise.Resolve(err) // TODO should be a union type for this
		return
	}
	mt.Promise.Resolve(exports) // TODO should be a union type for this
	return
}
