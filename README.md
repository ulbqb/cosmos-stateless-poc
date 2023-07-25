# cosmos-stateless-poc

This is PoC of stateless clients for cosmos chains. You can try the stateless client of Gaia.

## run gaia

```shell
$ git clone --branch v10.0.2 --depth=1 https://github.com/cosmos/gaia.git
$ cd gaia
$ go mod edit -replace github.com/cosmos/cosmos-sdk=github.com/ulbqb/cosmos-sdk@v0.45.17-0.20230717104443-b8f62f5fa8cf # Requires SDK with special API
$ make install
$ # run and sync
```

## run stateless client

```shell
$ cd example/gaiasl
$ go build
$ ./gaiasl -basedir ./tmp -rpc http://localhost:26657 -height 16182260 -hash 9DA91A055F29937A06AC8A15CD8D385C8BAD27AA148F9477B4F16AAFFD6AC5C5
E63D25384DEF73D71F096AEC15C35411B5DB5A50AFF8B58D259E1B8CDF5EE3C9 # next app hash by stateless execution
$ curl https://rpc-cosmoshub-ia.cosmosia.notional.ventures/block?height=16182261 | jq -r .result.block.header.app_has
E63D25384DEF73D71F096AEC15C35411B5DB5A50AFF8B58D259E1B8CDF5EE3C9 # next app hash by full node
```

## Implementation
- https://github.com/ulbqb/iavl/tree/v0.19.5-stateless-dev
    - Add witness tree
- https://github.com/ulbqb/cosmos-sdk/tree/v0.45.16-ics-stateless-dev
    - Add tree nodes getting API
    - Add witness tree
