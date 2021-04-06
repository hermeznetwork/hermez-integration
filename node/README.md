## Running a [node](https://github.com/hermeznetwork/hermez-node)

### Binaries

The latest binaries can be found on the Github [release page](https://github.com/hermeznetwork/hermez-node/releases).

- goos: `linux/darwin`;
- goarch: `amd64`;

### From Source

The binaries can be built from the source using the make commands (Golang version 1.14 or greater is required):

- Get the source: `go get github.com/hermeznetwork/hermez-node`;
- Run `make build`;
- The binaries are located into the `bin/` folder;
- Check the version `/bin/node version`;

### Run as a Synchronizer

Running the node requires a PostgreSQL database. 

For local development proposes the PostgreSQL with docker can be easily run (don't forget to change the password!):
```shell
$ docker run --rm --name hermez-db -p 5432:5432 -e POSTGRES_DB=hermez -e POSTGRES_USER=hermez -e POSTGRES_PASSWORD="<POSTGRES_PASSWORD>" -d postgres
```

After spinning up the database, you must edit the following parameters into the [`cfg.buidler.toml`](https://github.com/hermeznetwork/hermez-integration/blob/master/node/cfg.buidler.toml) config file:

```yaml
[StateDB]
Path = "<STATE_DB_DIRECTORY>/statedb"

[PostgreSQL]
PortWrite     = <POSTGRES_PORT>
HostWrite     = "<POSTGRES_HOST>"
UserWrite     = "hermez"
PasswordWrite = "<POSTGRES_PASSWORD>"
NameWrite     = "hermez"

[Web3]
URL = "<GETH_RPC_URL>"
```

The other values are already configured for the [Rinkeby testnet](https://docs.hermez.io/#/users/testnet). For the Mainnet usage, you must change the contract addresses for the [Hermez Mainnet](https://docs.hermez.io/#/users/mainnet).

There is more information about the config file parameters into [cli/node/README.md](https://github.com/hermeznetwork/hermez-node/blob/develop/cli/node/README.md).

After setting the config, you can build and run the Hermez Node as a synchronizer:
```shell
$ bin/node run --mode sync --cfg cfg.buidler.toml
```

The description of the config parameters can be found [here](https://github.com/hermeznetwork/hermez-node/blob/master/config/config.go).
