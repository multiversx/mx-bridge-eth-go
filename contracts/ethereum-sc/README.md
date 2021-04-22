# Run locally

Install Hardhat:
`npm install --save-dev hardhat`

Run Ganache CLI (or use the Ganache UI)
`ganache-cli  --mnemonic "miss guide shadow quiz moral custom collect adjust kiwi husband hope include"`
Make sure the port that Ganache listes to is the same with the one that's configured in `hardhat.config.js` under the `ganache` network.

Deploy and setup smart contracts for the Bridge
`npx hardhat run scripts/deploy.js --network ganache`

Make a Deposit
`npx hardhat run scripts/deposit-afc-to-safe.js --network ganache`

