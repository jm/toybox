![toyboxes](/docs/images/toyboxes.png)

# toyboxes

A dependency management system for Lua on the [Playdate](https://play.date).

### Why not just use Lua Rocks or Git submodules?

Lua Rocks, unfortunately, isn't compatible with the Playdate implementation
of Lua.  Playdate uses `import` to pull in files and libraries at _compile 
time_, whereas Lua Rocks (and mainline Lua) uses `require` that pulls in
things at runtime.  The distinction matters and probably has a lot to do
with running on embedded hardware, but suffice to say that while Playdate
supports nearly all of Lua, it does not support the package ecosystem.

Git submodules are a fine way to manage repositories that you just want to
clone down and update.  Using `toybox` gives you that ability as well as the
ability to manage transitive dependencies (and their proper, non-conflicting
versions) in those repositories.

## Usage

A toybox dependency can either be specifically tailored to use with toybox,
or a simple Git repository that follows most recommended Playdate development
patterns.  I've tried to make it accommodate what I see most folks doing and
Panic recommending, so it should require little to no work to setup a library
to work with `toybox`.

### Adding a dependency

To start, you'll need a file named `Boxfile` to your root folder.  A `Boxfile`
is a simple JSON file that is a single object document.  The object's keys 
are repo identifiers (`<GitHub username>/<repository name>`) and the values 
are the version requirements.  You can add a dependency using the `add` command
(and remove them using the `remove` command) like so:

```
toybox add Nikaoto/deep
```

Now your `Boxfile` should look like this:

```
{
	"Nikaoto/deep": "default"
}
```

The `default` version requirement will grab whatever the default branch for that
repository is (usually `master` or `main`).  You can also specify that using `*`
or the actual branch name.  You can optionally just edit your `Boxfile` like a 
normal text file to add the configuration, but the commands make it easy!  

For varying version requirements (for example, "anything newer than 1.0" or 
"greater than 2.0 but less than 3.0"), you can specify those in the same way.
Let's pretend we wanted any version newer than `1.0` of the `deep` library:

```
{
	"Nikaoto/deep": "> 1.0"
}
```

Versions in `toybox` are specified by Git tags on the repository, so if a library
owner wanted to publish a version, they simply have to tag it with a
[semver version](https://semver.org) number like `1.0` or `2.4.1` or `1.3.6beta` 
or some such.  The `toybox` client will parse those and then check the version
constraints specified in the various `Boxfiles` it resolves when installing to find
the best version.

### Installing the dependencies

Once you have a `Boxfile` setup, you simply run:

```
toybox install
```

This command will resolve and install the needed dependencies specified in your 
`Boxfile` (and the `Boxfiles` of your dependencies, and their dependencies, and
so on).  The packages are downloaded to `source/libraries` and namespaced by
toybox name.  Then a single import file is generated at `source/toyboxes.lua`:

```
import("libraries/Nikaoto/deep/deep")
```

So to import all of your toyboxes to your game, simply add:

```
import "toyboxes"
```

...and they should be available to use.

#### Dependencies of dependencies

Toybox will also handle getting the dependencies of your dependencies.  So let's
pretend that the `Nikaoto/deep` library depended on another library named `jm/geometry`.
In the `Boxfile` in the `Nikaoto/deep` repository, it would look something like this:

```
{
	"jm/geometry": "default"
}
```

Now if you ran `toybox install`, the output would look something like this:

```
🧸toybox v.0.1
Loading Boxfile...
Resolving dependencies...
Installing
Fetching jm/geometry@main
Fetching Nikaoto/deep@master
Writing import file
```

And the import file might look like (in `toyboxes.lua`):

```
import("libraries/jm/geometry/main")
import("libraries/Nikaoto/deep/deep")
```

Noting that it imports the dependency library before the `deep` library
(though that doesn't matter currently due to the way the Playdate SDK
compiles things, it's still a good future-proofing just in case!).  

Toybox will resolve these dependencies in an infinitely deep graph (i.e.,
it will get dependency of dependencies of dependencies of dependencies...),
so you only need to focus on your immediate dependencies and let `toybox`
take care of the rest.

### Other commands

The `toybox` binary has a few other subcommands.

#### Raising GitHub API limits by logging in

If you're doing a lot of dependency changes, you might run up against
the GitHub API's rate limiting on unauthenticated requests.  If this
happens, you'll start seeing the request status for things like getting
a list of versions returning a `403` status rather than `200`.  To fix
that, you'll need to provide a [GitHub personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).

Following the instructions in the above link, you can generate a token
that you'll need to copy and paste into the prompts when asked:

```
toybox login
```

Toybox will ask for your username and the token.  Once provided, the token will
be passed with each request.  Authenticated API limits are much higher and 
shouldn't cause you any issues going forward.

**Note:**  When generating the token, it's best if you choose a sensible
expiration (60 days or so?) and _only_ add the `repo:public_repo` scope.
That way if your token somehow gets compromised, the only thing the person
who has it can do it read public repositories (which is all `toybox` needs
to do unless you have private dependencies in private repos).

#### Generating a pre-wired game project

To generate a new Playdate project pre-wired for `toybox` and with a few 
extra goodies (project structure, `pdxinfo`, `.gitignore`, `Makefile`, 
etc.), use the `generate` command:

```
toybox generate path/to/your/project
```

This command will drop a new Playdate project in the given path (named
for the last part of the path), that you can immediately run `make` and
run in the Simulator.

#### Managing dependencies in your `Boxfile`

You can add and remove dependencies manually if you'd like, but there are
also convenience commands for doing that:

```
toybox add owner/dependency
toybox remove owner/dependency
```

These commands will add or remove a dependency from your `Boxfile` and 
resolve/install the list again.

If you want to update a single dependency's version (either because you
changed it manually or because there is a newer one available), you can
run:

```
toybox update owner/dependency
```

Running this command will find the best version to resolve with the 
provided version constraints (including the changed version for the
dependency you're updating).

#### Getting information

To find out the current `toybox` version, simply run `toybox version`.

To see the current list of dependencies, run `toybox info`.

To get more help, run `toybox help` followed by a subcommand.  For
example, `toybox help install`.

## Creating a toybox

Creating a library compatible with `toybox` is actually really simple
since I tried to make `toybox` understand conventions and norms that
the PlayDate dev community is already using.

### Setting up the code

Ideally, your code will be contained in a `source` folder with a 
`import.lua` file that is the entrypoint for your library.
Realistically, though, your code simply needs to have one of the
following files to be properly imported by `toybox`:

* `/source/import.lua`
* `/source/main.lua`
* `/source/<your toybox name>.lua` (e.g., `jm/Geometry/source/Geometry.lua`)
* `/import.lua`
* `/main.lua`
* `/<your toybox name>.lua` (e.g., `jm/Geometry/source/Geometry.lua`)

That's about it.  Once someone imports your package using `toybox`,
the code will be available after they import the `toyboxes.lua` file.

If your toybox depends on other toyboxes, you can add a `Boxfile`
to your library and `toybox` will resolve and import those into 
your users' downstream code as well.  So, for example, if `jm/Geometry`
depends on `you/Things` and a user added the `Geometry` library to their
project's `Boxfile`, the dependency list shown by `toybox info` would 
list both of xthose dependencies as being in their toybox bundle.

### Making a release

Folks can always install your toybox from `master` or `main` or whatever
the default branch is for your repository.  They can install that version,
and when new commits are added, they can run `toybox update` to get the
newest code.  Ideally, though, your library would be versioned so users
can better manage its impact on their projecrts.

To make a versioned release, you need to tag a commit in your Git
repository and push that tag to GitHub.  You can do this from the command
line with [these instructions](https://git-scm.com/book/en/v2/Git-Basics-Tagging)
or you can do it in the GitHub UI with their[fancy Releases feature](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases).

The tag's name needs to follow [Semantic Versioning](https://semver.org).  So for example, good
tag names would be something like:

* `v8.0`
* `v0.2.0`
* `0.1.1beta`

You can name your releases whatever you'd like, but the version resolution
might get interesting if you don't [follow semantic versioning](https://semver.org). 😬

## Contributing

[File issues for any bugs or feature requests!](https://github.com/jm/toybox/issues)

If you want to contribute code, please create a feature branch and [submit
a pull request](https://github.com/jm/toybox/pulls).
