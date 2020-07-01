# Developer notes

1. [Makefile](#makefile)
2. [OBS packaging](#obs-packaging)


## Makefile

Most development tasks can be accomplished via [make](../Makefile).

For starters, you can run the default target with just `make`.

The default target will clean, analyse, test and build the amd64 binary into the `build/bin` directory.

You can also cross-compile to the various architectures we support with `make build-all`.


## OBS Packaging

The CI will automatically publish GitHub releases to SUSE's Open Build Service: to perform a new release, just publish a new GH release. Tags must always follow the [SemVer](https://semver.org/) scheme.

If you wish to produce an OBS working directory locally, having configured [`osc`](https://en.opensuse.org/openSUSE:OSC) already, you can run:
```
make exporter-obs-workdir
```
This will checkout the OBS project and prepare a new OBS commit in the `build/obs` directory.

You can use the `OSB_PROJECT`, `REPOSITORY`, `VERSION` and `REVISION` environment variables to change the behaviour of OBS-related make targets.

By default, the current Git working directory is used to infer the values of `VERSION` and `REVISION`, which are used by OBS source services to generate a compressed archive of the sources.  

For example, if you were on a feature branch of your own fork, you may want to change these variables, so:
```bash
git checkout feature/xyz
git push johndoe feature/xyz # don't forget to push changes your own fork remote
export OBS_PROJECT=home:JohnDoe
export REPOSITORY=johndoe/prometheus-ha_cluster_exporter
make exporter-obs-workdir
``` 
will prepare to commit on OBS into `home:JohnDoe/prometheus-ha_cluster_exporter` by checking out the `feature/xyz` branch from `github.com/johndoe/my_forked_repo`.

At last, to actually perform the commit into OBS, run: 
```bash
make exporter-obs-commit
```

Note that that actual continuously deployed releases also involve an intermediate step that updates the changelog automatically with the markdown text of the GitHub release.
