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
ability to manage transient dependencies (and their proper, non-conflicting
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

And the import file might look like:

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

...

#### Getting information

...

## Creating a toybox

...

### Setting up the code

...

### Making a release

...

## Contributing

...
