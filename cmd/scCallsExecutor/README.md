# MultiversX - EVM-compatible chains bridge SC calls executor

This tool allows the execution of the SC calls that can be made in the direction EVM-compatible chain -> MultiversX.
The key provided to this tool will spend the gas limit required to execute the SC call. In case the SC call fails, 
a refund operation is generated and saved inside the Bridge Proxy smart-contract. This tool can then trigger the whole 
refund operation spending gas from the same provided key.

## Feature list
- [x] Monitor any number of Bridge Proxy smart-contracts on a network;
- [x] The network can be provided, along with other network-related parameters; 
- [x] There is a filter that can be used to allow the sending of transactions only if they match a certain criteria:
  - [x] Matching after an EVM-chain address
  - [x] Matching after a MultiversX address
  - [x] Matching after a token (MultiversX ESDT definition)
  - [x] Allow or disallow one or more definitions. Disallow takes precedence after the allow definition
- [x] If a SC call or refund transaction failed, there are configurable cool-down periods before the transactions can be re-sent;
- [x] Scripts & installation support
    - [x] Added scripts for easy setup & upgrade
    - [x] Added Docker image build & scripts

## Installation

You can choose to run this tool either in a Docker on in a systemd service.

### Initial setup (valid for all types of installation)

Although it's possible, it is not recommended to run the application as `root`. For that reason, a new user is required to be created.
For example, the following script creates a new user called `ubuntu`. This script can be run as `root`.

```bash
# host update/upgrade
apt-get update
apt-get upgrade
apt autoremove

adduser ubuntu
# set a long password
usermod -aG sudo ubuntu
echo 'StrictHostKeyChecking=no' >> /etc/ssh/ssh_config

visudo   
# add this line:
ubuntu  ALL=(ALL) NOPASSWD:ALL
# save & exit

sudo su ubuntu
sudo visudo -f /etc/sudoers.d/myOverrides
# add this line:
ubuntu  ALL=(ALL) NOPASSWD:ALL
# save & exit
```

### Variant A: how to set up using Docker

You need to have [Docker](https://docs.docker.com/engine/install/) installed on your machine.

Copy the template configs from the `example` directory:
```bash
cd
git clone https://github.com/multiversx/mx-bridge-eth-go
cd mx-bridge-eth-go
cp ./cmd/scCallsExecutor/config/config.toml.example ./cmd/scCallsExecutor/config/config.toml
mkdir ./cmd/scCallsExecutor/keys
```

Customize your global tool's config file from the `./cmd/scCallsExecutor/config/config.toml`
Then add a MultiversX .pem private key in `./cmd/scCallsExecutor/keys`, it will be used to pay the gas for the SC calls

Fetch the image & start it, using the docker-compose.yml file:
```bash
cd docker
sudo docker compose -f ./scCallsExecutor-docker-compose.yml up -d sc-calls-executor
```

You're ready 🚀

### Variant B: how to set up using regular bash scripts (using service file)

#### 1. Repo clone & scripts init

```bash
cd
git clone https://github.com/multiversx/mx-bridge-eth-go
cd ~/mx-bridge-eth-go/scripts/scCallsExecutor
# the following init call will create ~/mx-bridge-eth-go/scripts/scCallsExecutor/config/local.cfg file
# and will copy the configs from ~/mx-bridge-eth-go/cmd/scCallsExecutor/config/config.toml.example to ~/mx-bridge-eth-go/cmd/scCallsExecutor/config/config.toml
# to avoid github pull problems
./script.sh init
cd config
# edit the local.cfg file for the scripts setup
nano local.cfg
```

#### 2. local.cfg configuration

The generated local.cfg file contains the following lines:

```bash
#!/bin/bash
set -e

CUSTOM_HOME=/home/ubuntu
CUSTOM_USER=ubuntu
GITHUBTOKEN=""
MONITOR_EXTRA_FLAGS=""

#Allow user to override the current version of the monitor
OVERRIDE_VER=""
```

The `CUSTOM_HOME` and `CUSTOM_USER` will need to be changed if the current user is not `ubuntu`. 
To easily figure out the current user, the bash command `whoami` can be used.

It is strongly recommended to use a GitHub access token because the scripts consume the GitHub APIs and
throttling might occur without the access token.

The `EXTRA_FLAGS` can contain extra flags to be called whenever the application is started. 
The complete list of the cli commands can be found [here](./cmd/scCallsExecutor/CLI.md) 

The `OVERRIDE_VER` can be used during testing to manually specify an override tag/branch that will be used when building 
the application. If left empty, the upgrade process will automatically fetch and use the latest release.

#### 3. Install

After the `local.cfg` configuration step, the scripts can now install the application.
```bash
cd ~/mx-bridge-eth-go/scripts/scCallsExecutor
./script.sh install
```

#### 4. Application config

After the application has been installed, it is now time to configure it.
For this, you should edit the `config.toml` and add the `multiversx.pem` file that will contain the private key that has 
access to the funds used to pay the gas in the SC call transaction.

The scripts init step already created the initial `config.toml` file that will be ignored by the future GitHub's `pull` commands.

Configuring the **config.toml** file:

This file contains the general application configuration file.
```toml
[General]
    ScProxyBech32Addresses = [
        "erd1qqqqqqqqqqqqqpgqnef5f5aq32d63kljld8w5vnvz4gk5sy9hrrq2ld08s",
    ]
    NetworkAddress = "http://127.0.0.1:8085"
    ProxyMaxNoncesDelta = 7
    ProxyFinalityCheck = true
    ProxyCacherExpirationSeconds = 600
    ProxyRestAPIEntityType = "proxy"
    IntervalToResendTxsInSeconds = 60
    PrivateKeyFile = "keys/multiversx.pem"

[ScCallsExecutor]
    ExtraGasToExecute = 60000000 # this value allow the SC calls without provided gas limit to be refunded
    MaxGasLimitToUse = 249999999 # this is a safe max gas limit to use both intra-shard & cross-shard
    GasLimitForOutOfGasTransactions = 30000000 # this value will be used when a transaction specified a gas limit > 249999999
    PollingIntervalInMillis = 6000
    TTLForFailedRefundIdInSeconds = 3600

[RefundExecutor]
    GasToExecute = 30000000
    PollingIntervalInMillis = 6000
    TTLForFailedRefundIdInSeconds = 86400

[Filter]
    AllowedEthAddresses = ["*"]   # execute SC calls from all ETH addresses
    AllowedMvxAddresses = ["*"]   # execute SC calls to all MvX contracts
    AllowedTokens = ["*"]         # execute SC calls for all tokens

[Logs]
    LogFileLifeSpanInSec = 86400 # 24h
    LogFileLifeSpanInMB = 1024 # 1GB

[TransactionChecks]
    TimeInSecondsBetweenChecks = 6     # the number of seconds to recheck the status of the transaction
    ExecutionTimeoutInSeconds  = 120   # the number of seconds reserved for each execution to complete
```

* The `General` section:
  - The `ScProxyBech32Addresses` will contain all individual Bridge Proxy contracts that the tool will monitor.
  - The `NetworkAddress` will contain the proxy (gateway) address. Example: `https://devnet-gateway.multiversx.com` or
`https://gateway.multiversx.com` or a simple node's address. In case a simple node is used, the setting 
`ProxyRestAPIEntityType` should be changed to `observer`.
  - `ProxyMaxNoncesDelta` and `ProxyFinalityCheck` define the finality setting when querying the proxy address.
  - `ProxyCacherExpirationSeconds` sets how many seconds will the tool use the chain's settings like the chain ID or 
minimum gas price before they are queried again from the proxy.
  - `ProxyRestAPIEntityType` should have either the `proxy` or `observer` values. 
  - `IntervalToResendTxsInSeconds` sets the seconds that a transaction re-send process is triggered.
  - `PrivateKeyFile` sets the default path for the key that this tool uses to call smart-contracts.

* The `ScCallsExecutor` section:
  - `ExtraGasToExecute` sets the additional gas that will be used when doing the SC call, added to the gas limit 
provided in the swap transfer.
  - `MaxGasLimitToUse` the gas limit will be trimmed to this value as to avoid the case when the transaction will be 
discarded due to a very high gas limit set on it.
  - `GasLimitForOutOfGasTransactions` if a swap deposit with SC call has an over the limit gas limit, this value will 
be used instead so the transaction will be marked as failed and able to be refunded.
  - `PollingIntervalInMillis` the number of milliseconds between queries on the Bridge Proxy contracts.
  - `TTLForFailedRefundIdInSeconds` represents the number of seconds a SC call ID will be in cool-down period if 
the transaction fails in such a way that is neither succeeded nor failed. This is an extreme case, and, without a 
considerable high value for this parameter, funds drainage on the provided key is possible.

* The `RefundExecutor` section:
  - `GasToExecute` sets the gas to be used when trying to execute the refund transaction. A refund transaction is a 
failed SC call transaction that will return the amount sent - fee to the original sender of the swap transfer.
  - `PollingIntervalInMillis` the number of milliseconds between queries on the Bridge Proxy contracts.
  - `TTLForFailedRefundIdInSeconds` represents the number of seconds a refund ID will be in cool-down period if
the transaction fails. A refund transaction should not normally fail but this parameter along with the cool-down 
mechanism can prevent any funds drainage on the provided key.

* The `Filter` section:
  
This section allows executing only the swap transfer that match a certain or cumulated criteria.

  - `AllowedEthAddresses` defines the list of the allowed senders from the EVM-compatible chain. 
The `*` value means all addresses are allowed.
  - `AllowedMvxAddresses` defines the list of the allowed destination SC addresses from MultiversX.
The `*` value means all addresses are allowed.
  - `AllowedTokens` defines the list of the allowed MultiversX ESDTs.
The `*` value means all tokens are allowed.
  - `DeniedEthAddresses` defines the list of the specifically denied senders from the EVM-compatible chain.
  - `DeniedMvxAddresses` defines the list of the specifically denied destination SC addresses from MultiversX.
  - `DeniedTokens` defines the list of the specifically denied MultiversX ESDTs.

If an element is found in a denied list, the tool will not consider the swap transfer that contains the element even if 
that element satisfies the allowed list. To be considered, a swap should have the sender, destination and token elements 
capable of satisfying the allowed lists.

#### 5. Application start

After editing the required config files, the application can be started.
```bash
cd ~/mx-bridge-eth-go/scripts/scCallsExecutor
./script.sh start
```

#### 6. Backup and upgrade

It is a good practice to save the config.toml, multiversx.pem and the local.cfg files somewhere else just in case the 
application is cleaned up accidentally.
The upgrade call for the monitor app is done through this command:
```bash
cd ~/mx-bridge-eth-go/scripts/scCallsExecutor
./script.sh upgrade
```

#### 7. Uninstalling

The application can be removed by executing the following script:
```bash
cd ~/mx-bridge-eth-go/scripts/scCallsExecutor
./script.sh cleanup
```

### Troubleshooting

If the application fails to start (maybe there is a bad config in the config.toml file), the following command can be issued:
```bash
sudo journalctl -f -u mx-bridge-sc-calls-executor.service
```

Also, if the application misbehaves, the logs can be retrieved by using this command:
```bash
cd ~/mx-bridge-eth-go/scripts/scCallsExecutor
./script.sh get_logs
```

If the application crashes, and you have followed the installation via Docker, the command to retrieve the logs is as follows:
```bash
sudo docker logs -f sc-calls-executor
```
