# MultiversX<->Eth Bridge
The relayer code implemented in go that uses the smart contracts and powers the bridge between MultiversX and Ethereum.

Smart contracts for both blockchains:
- https://github.com/multiversx/mx-bridge-eth-sc-rs
- https://github.com/multiversx/mx-bridge-eth-sc-sol

## Installation and running for the relayer

### Step 1: install & configure go:
The installation of go should proceed as shown in official golang installation guide https://golang.org/doc/install . In order to run the node, minimum golang version should be 1.14.7.

### Step 2: clone the repository and build the binaries:
The `main` branch is the one to use

### Step 3: configure the relay
Checkout `config.toml.example` for all the configuration needed:

### Step 4: monitoring your relayer node
After your node is up and running. You can use relayer's api routes to monitor the existing metrics.
For the documentation and how to setup swagger. Go to [README.md](api/swagger/README.md)


## Contribution
Thank you for considering to help out with the source code! We welcome contributions from anyone on the internet, and are grateful for even the smallest of fixes to Elrond!

If you'd like to contribute to Elrond, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base. If you wish to submit more complex changes though, please check up with the core developers first on our [telegram channel](https://t.me/ElrondNetwork) to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much lighter as well as our review and merge procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

- Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting) guidelines.
- Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
- Pull requests need to be based on and opened against the master branch.
- Commit messages should be prefixed with the package(s) they modify.
    - E.g. "core/indexer: fixed a typo"

Please see the [documentation](https://docs.elrond.com/) for more details on the MultiversX project.
