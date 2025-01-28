
# Bridge SC calls Executor CLI

The **MultiversX Bridge SC calls executor** exposes the following Command Line Interface:

```
$ scCallsExecutor --help

NAME:
   SC calls executor CLI app - This is the entry point for the module that periodically tries to execute SC calls

USAGE:
   scCallsExecutor [global options] command [command options] [arguments...]

VERSION:
   undefined/go1.20.7/linux-amd64/03d1f4fa88

AUTHOR:
   The MultiversX Team <contact@multiversx.com>

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --working-directory directory          This flag specifies the directory where the node will store databases and logs.
   --log-level level(s)                   This flag specifies the logger level(s). It can contain multiple comma-separated values. For example, if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG log level. (default: "*:DEBUG")
   --disable-ansi-color                   Boolean option for disabling ANSI colors in the logging system.
   --config [path]                        The [path] for the main configuration file. This TOML file contains the main configurations such as monitored SC, gateway URL, timings and so on (default: "config/config.toml")
   --log-save                             Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.
   --log-logger-name                      Boolean option for logger name in the logs.
   --profile-mode                         Boolean option for enabling the profiling mode. If set, the /debug/pprof routes will be available on the node for profiling the application.
   --rest-api-interface address and port  The interface address and port to which the REST API will attempt to bind. To bind to all available interfaces, set this flag to :8080 (default: "localhost:8080")
   --network-address value                The network address (gateway) to be used. Example: 'https://testnet-explorer.multiversx.com'
   --private-key-file value               The MultiversX private key file used to issue transaction for the SC calls

```

