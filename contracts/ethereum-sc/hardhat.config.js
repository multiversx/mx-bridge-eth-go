const { task } = require("hardhat/config");

require("@nomiclabs/hardhat-waffle");
require("@nomiclabs/hardhat-solhint");
require("hardhat-watcher");
require("hardhat-gas-reporter");


// This is a sample Hardhat task. To learn how to create your own go to
// https://hardhat.org/guides/create-task.html
task("accounts", "Prints the list of accounts", async () => {
  const accounts = await ethers.getSigners();

  for (const account of accounts) {
    console.log(account.address);
  }
});

task("add-to-whitelist", "Whitelists a new address in the bridge. Requires setup.config.json to be present (created with the deploy script)")
  .addParam("address", "Address of the ERC20 token to be whitelisted")
  .setAction(async taskArgs => {
    const tokenAddress = taskArgs.address;
    [adminWallet] = await ethers.getSigners();
    const fs = require('fs');
    config = JSON.parse(fs.readFileSync('setup.config.json', 'utf8'));
    const safeAddress = config["erc20Safe"];
    const safeContractFactory = await ethers.getContractFactory("ERC20Safe");
    const safe = await safeContractFactory.attach(safeAddress).connect(adminWallet);
    await safe.whitelistToken(tokenAddress);
    console.log("Token whitelisted: ", tokenAddress);
  })

task("set-quorum", "Updates the quorum on the Bridge contract")
  .addParam("newQuorumSize", "Integer representing the quorum for a transfer to be considered valid")
  .setAction(async taskArgs => {
    const newQuorumSize = taskArgs.newQuorumSize;
    [adminWallet] = await ethers.getSigners();
    const fs = require('fs');
    config = JSON.parse(fs.readFileSync('setup.config.json', 'utf8'));
    const bridgeAddress = config["bridge"];
    const bridgeContractFactory = await ethers.getContractFactory("Bridge");
    const bridge = await bridgeContractFactory.attach(bridgeAddress).connect(adminWallet);
    result = await bridge.setQuorum(newQuorumSize);
    console.log("Quorum updated: ", newQuorumSize);
  })

// You need to export an object to set up your config
// Go to https://hardhat.org/config/ to learn more

/**
 * @type import('hardhat/config').HardhatUserConfig
 */
module.exports = {
  solidity: "0.8.5",
  networks: {
    ganache: {
      url: 'http://127.0.0.1:8545'
    },
    rinkeby: {
      url: "https://rinkeby.infura.io/v3/df34d380f59e469c97f1dab44199bca6",
      accounts: {
        mnemonic: "industry layer bird test junk shadow visa lottery human spatial pact balcony"
      }
    }
  },
  watcher: {
    test: {
      tasks: [{ command: 'test', params: { testFiles: ['{path}'] } }],
      files: ['./test/**/*'],
      verbose: true
    },
    compilation: {
      tasks: ["compile"],
      files: ["./contracts"],
      verbose: true,
    },
    ci: {
      tasks: ["clean", { command: "compile", params: { quiet: true } }, { command: "test", params: { noCompile: true, testFiles: ["testfile.ts"] } }],
    }
  },
  gasReporter: {
    currency: 'USD',
    gasPrice: 10,
    coinmarketcap: '26043cba-19e3-4a70-8575-916adb54fa12'
  }
};

