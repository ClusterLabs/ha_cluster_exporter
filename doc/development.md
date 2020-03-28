# Developer notes

1. [Makefile](#makefile)
2. [OBS packaging](#obs-packaging)


## Makefile

Most development tasks can be accomplished via [make](../Makefile).

For starters, you can run the default target with just `make`.

The default target will clean, analyse, test and build the amd64 binary into the `build/bin` directory.

You can also cross-compile to the various architectures we support with `make build-all`.


## OBS Packaging

The CI will automatically publish GitHub releases to SUSE's Open Build Service: to perform a new release, just publish a new GH release or push a git tag. Tags must always follow the [SemVer](https://semver.org/) scheme.

If you wish to produce an OBS working directory locally, having configured [`osc`](https://en.opensuse.org/openSUSE:OSC) already, you can run:
```
make obs-workdir
```
This will checkout the OBS project and prepare a new OBS commit in the `build/obs` directory.

Note that, by default, the current Git working directory HEAD reference is used to download the sources from the remote, so this reference must have been pushed beforehand.
  
You can use the `OSB_PROJECT`, `OBS_PACKAGE`, `REPOSITORY` and `VERSION` environment variables to change the behaviour of OBS-related make targets.

For example, if you were on a feature branch of your own fork, you may want to change these variables, so:
```bash
git push feature/yxz # don't forget to make changes remotely available
export OBS_PROJECT=home:JohnDoe
export OBS_PACKAGE=my_project_branch
export REPOSITORY=johndoe/my_forked_repo
export VERSION=feature/yxz
make obs-workdir
``` 
will prepare to commit on OBS into `home:JohnDoe/my_project_branch` by checking out the branch `feature/yxz` from `github.com/johndoe/my_forked_repo`.

At last, to actually perform the commit into OBS, run `make obs-commit`. 
