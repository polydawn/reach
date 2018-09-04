hacking stellar
===============

code layout
-----------

- `gadget/*/*` -- libraries and anything that you fire of and expect to return.
- `actors/*/*` -- features that use gadgets in some sort of daemonized mode (they have channels and need supervisors).
- `app/*/*` -- actual get-concrete-thing-done assemblies of gadgets and actors.
- `cmd/*` -- the main method that turns all the apps into a usable CLI.
