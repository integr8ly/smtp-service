# RHMI SMTP Tooling and Service

This repo is intended to store the services and tooling related to the setup of
SMTP in an Integreatly/RHMI cluster.

## CLI

This repo contains a CLI which can be used to create SendGrid sub users and API
keys for those sub users, that can be used with RHMI clusters.

The intended pattern is that each RHMI cluster would have it's own sub user
which would then contain an API key for the RHMI cluster. The reasoning behind
this approach is that *SendGrid clusters only allow 100 API keys per user*.

### Building

To build the CLI, run from the root of this repo:

```
make build/cli
```

A binary will be created in the root directory of the repo, which can be run:

```
./cli
```


### How to use

To use the CLI, you must create the env var `SENDGRID_API_KEY`, with a SendGrid master account API key with at least
permissions to:

- Create and read sub users
- Create and read API keys
- Read IP addresses

To export the env var, run:

```
export SENDGRID_API_KEY=<mySendGridAPIKey>
```

#### Create a new API key for a cluster

To create a new API key for a cluster, run:

```
./cli create my_cluster_id
```

an OpenShift Secret will be output to stdout.

Note that the cluster name must also be a unique username is SendGrid.

#### Delete an API key for a cluster

To delete an API key for a cluster, run:

```
./cli delete my_cluster_id
```

This will simply delete the sub user associated with the SendGrid cluster.

#### Get an API key for a cluster

To get the name of the API key for a cluster, not the cluster itself, run:

```
./cli get my_cluster_id
```

This command is mainly useful to check if an API key exists for the cluster.

## Testing

To run unit tests, run:

```
make test/unit
```

## Releases

New binaries for a release tag will be created by [GoReleaser](https://goreleaser.com/) automatically.

To try out GoReleaser locally, it can be installed using `make setup/goreleaser`. 