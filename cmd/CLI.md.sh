#!/usr/bin/env bash

generate() {
    generateForSCCallsExecutor
}

generateForSCCallsExecutor() {
    HELP="
# Bridge SC calls Executor CLI

The **MultiversX Bridge SC calls executor** exposes the following Command Line Interface:
$(code)
\$ scCallsExecutor --help

$(./scCallsExecutor/scCallsExecutor --help | head -n -3)
$(code)
"
    echo "$HELP" > ./scCallsExecutor/CLI.md
}

code() {
    printf "\n\`\`\`\n"
}

generate
