Stellar
=======

Stellar is a core part of [the Timeless Stack](https://repeatr.io);
Stellar understands pipelines of computations, and generates formulas
for [Repeatr](https://github.com/polydawn/repeatr) to evaluate.

tl;dr: if you're looking at Repeatr and thinking
"wow, that's neat, but I don't want to copy-paste a bunch of hashes"...
Stellar is the thing that fixes that for you.


By Example
----------

### hellomodule

- You can see a stellar module here:
  [./examples/hellomodule/module.tl](./examples/hellomodule/module.tl).
- The imports which are `"catalog:x:y:z"` references are resolved into WareID hashes
  by looking at files in
  [./examples/hellomodule/.timeless/catalog](./examples/hellomodule/.timeless/catalog).
- You can see the expected output in the test fixtures here:
  [./examples/hellomodule/helloModule_test.go](./examples/hellomodule/helloModule_test.go).

This is a very basic example, and only has one "step" -- so it will only generate one formula.

This example is still interesting, because it demonstrates the use of catalogs to
solve the `human-readable-name => hash` usability issue.

### more examples

Look at the other directories in `./examples` :)

You'll find examples of more complex modules with multiple steps -- Stellar
will automatically resolve dependencies in these modules, and execute steps in
dependency order, feeding outputs of one step into inputs of the next.

You'll also find examples of modules which contain "submodules", which are
nested definitions that look like more modules, but can have `"parent:{localName}"`
references as their imports -- these give us a way to create reusable snippets
of pipeline which have locally scoped names (like a function), and thus we can
compose them without having to do name munging.
(You should still probably write Layer 3 helper functions to compose things, but
locally scoped names in the serial format help by taking the munging work out,
and also make the result clearer to read.)


API edges
---------

Stellar drives Repeatr around via JSON API.

Stellar is meant to be a user-facing CLI tool.  You can configure it with json config files.

Stellar json files (namely modules) are meant to be human-readable and human-writable.
However, they can also get quite verbose for large projects.
It is explicitly intended to be reasonable to generate module json with a higher level language.
(In other words, Stellar is "Layer 2" in the Timeless Stack model, and you should feel free to implement your own business logic constructs in "Layer 3" that generate module json.)

Stellar only performs *non-turing complete* operations (excluding what occurs inside containers, of course) -- that is, the stellar modules express a dependency DAG, but
there is not (and will not be) support for modules which generate more steps, etc.
Build this kind of feature at Layer 3.


Other things that deserve a mention in docs
-------------------------------------------

Stellar development is in a rapid iteration phase and PRs for documentation are more than welcome :)

- workspaces
- ingests -- git and pack and how to use them (well)
- ci mode
- catalog authoring and syncing tools
- relationship of "modules" to "replay instructions" in releases (they're the same!)
- "candidates mode" for distro-scale multi-module release coordination
