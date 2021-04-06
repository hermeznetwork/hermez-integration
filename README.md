# Hermez Integration Examples

Go examples for Hermez Network integration:

- Create BJJ wallets form a mnemonic;
- Calculate fee;
- Sign L2 transactions;
- Get the last batch;
- Get all transactions from a batch;
- Track transactions forged and in the pool;

## Developing

### Go version

The `hermez-integration` has been tested with go version 1.16

### Usage

Build and run the example binary:

```shell
$ make run
```

_This repository cannot be used as a go library. It's only examples of how to implement the integration in Go._

### Node 

The information about deploying a node as a synchronizer can be found [here](node/README.md)
