# open-custodial

An open source private key management system with modular integrations at the storage layer. **_Not for production use_**... yet 😉

At the moment, this project is focusing on Ethereum, secp256k1, and HSM's.

Each package is designed to be modular enough to plug and play in whichever way you feel is useful.

### `open_custodial/pkg/hsm`

An interface for interacting with an HSM using pkcs11 bindings. 

### `open_custodial/pkg/eth_hsm`

A library for doing common Ethereum operations with an HSM interface.

### `open_custodial/module/eth`

A server to invoke `eth_hsm` functionality via http requests.
