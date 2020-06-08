
The Relationship between Workspaces, Modules, and Version Control
-----------------------------------------------------------------

Every Module must be within a workspace to function.
(It's not possible to do import resolution, otherwise.)

A Workspace that contains no modules is valid.
(It might even have utility: for polinating catalogs, for example.)

A Version Control repo (such as a git repo) _typically_ contains one Workspace and one Module;
and _typically_ they are both in the root directory of the repo.
(This is the minimum necessary for self-contained testing.)
This arrangement usually features a very minimal Catalog embedded in the Workspace,
containing just enough info to satisfy the Module's import resolution;
because this is common, we make tool features specifically to support this (and keep it trim).

A Version Control repo with neither a Workspace nor a Module isn't using Timeless Stack.

A Version Control repo that contains a workspace and _many_ modules is perfectly valid.
(This might be a common sight in an organization that works with a monorepo, but has many projects, for example;
or at really large scale, this might even represent a "distro"!)

A Version Control repo with a Module but no Workspace is probably a bad idea
(and our tools will warn about this loudly, if they notice it) --
this means the Version Control repo is dependent on some external Workspace
in order for the Module to perform import resolution,
and that means reproducible resolve of imports is not tracked,
which is a bit helter-skelter and just not good practice.

A Version Control repo that contains Workspace configuration but no Modules
is a bit odd, but fine.
(If one wants to version control Catalog info, though, one usually just tracks those files;
there's no need to also track a set of Workspace configuration files just to keep the Catalog files company.)


Gathering Modules from Far Afield
---------------------------------

Modules must be within a Workspace; and more specifically,
the Module root path must be within the Workspace's root path.

You still don't want to consider it exactly 1:1, though.
Module configuration is designed so that it's still usable (without editing!)
in more than one Workspace.

There's three user stories where this is relevant: reparenting, cloning, and symlinks.


### reparenting

Modules are fairly explicit about how they import information from their surroundings,
but the effects they _export_ can vary based on the workspace...
and some operations less read-only operations like import updating can also be sensitive to workspace.

Consider the following filesystem:

```
/universe/
/universe/.timeless/    # a workspace!
/universe/organization/
/universe/organization/.timeless/   # another workspace!
/universe/organization/project/
/universe/organization/project/module.tl   # this is a module!
/universe/organization/project/.git/        # it happens to be version controlled
/universe/organization/project/foobar.code   # dunno what this is (but the module imports tools that do, presumably)
/universe/organization/project/.timeless/   # the project's personal workspace!
```

These three commands will all do similar, but not quite the same, things:

- `cd /universe && reach emerge ./organization/project`
- `cd /universe/organization && reach emerge ./project`
- `cd /universe/organization/project && reach emerge .`

If all three workspaces have the same Catalog contents and the same module name configuration, then all of those actions are effectively the same.

However, if you're doing import update probing, and one of those workspaces has newer Catalog contents than the others, that'll act differently;
if you're generating sagas of Candidates, those will of course be stored in the workspace you're operating on;
and last of all (and most interestingly)... workspaces can have name overrides for modules at any path inside them,
meaning the _module name_ for that "project" folder could be different for any those three commands.


### cloning

Let's continue the example earlier, but now suppose the project repo is cloned by someone else.

Their filesystem now looks like this:

```
/bobverse/
/bobverse/.timeless/    # a workspace!
/bobverse/bobcorp/
/bobverse/bobcorp/.timeless/   # another workspace!
/bobverse/bobcorp/project/
/bobverse/bobcorp/project/module.tl   # this is a module!
/bobverse/bobcorp/project/.git/        # it happens to be version controlled
/bobverse/bobcorp/project/foobar.code   # dunno what this is (but the module imports tools that do, presumably)
/bobverse/bobcorp/project/.timeless/   # the project's personal workspace!
```

_This should be fine_... as long as all of the imports needed by the project can be resolved.
And since the project repo contains its own self-contained workspace (which should have those imports in its catalog),
that should pretty much already be settled.
Cloning the repo into a new and different context should "just work".

This should be also continue to be fine
_even if the "bobverse" or "bobcorp" workspaces use moduleName override directives_ from one of the enclosing workspaces
to change the name the project annotates its logs and artifacts with.


### symlinks

I don't really know what good you'd get up to using symlinks...
But, it should work.

(If you're doing this to make a module belong to more than one workspace,
consider what it's going to be like to maintain it.
You might end up saner if you just clone more -- otherwise the symlink may
create scenarios where the two separate workspaces each attempting to reconile
and verify the module's contents (or do operation like import updates) might
start prompting polinations in a way that could get confusing to manage.)
