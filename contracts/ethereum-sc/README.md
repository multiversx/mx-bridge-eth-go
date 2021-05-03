# Run locally

1. Install Hardhat:

`npm install --save-dev hardhat`

2. Run Ganache CLI (or use the Ganache UI)

`ganache-cli  --mnemonic "miss guide shadow quiz moral custom collect adjust kiwi husband hope include"`

Make sure the url that Ganache listes to is the same with the one that's configured in `hardhat.config.js` under the `ganache` network.

3. Deploy and setup smart contracts for the Bridge
`npx hardhat run scripts/deploy.js --network ganache`

4. Make a Deposit
`npx hardhat run scripts/deposit-afc-to-safe.js --network ganache`

# Develop

Use watcher for easy TDD.

`npx hardhat watch test` to watch any files on test files and have that test file run.
Combine this with `it.only('test')` to very quickly run the modified test.