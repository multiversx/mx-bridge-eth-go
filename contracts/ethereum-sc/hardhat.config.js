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
    await safe.whitelistToken(tokenAddress, 0);
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

task("deploy", "Deploys ERC20Safe and the Bridge contract")
  .addParam("relayerAddresses", "JSON Array containing all relayer addresses to be added when the Bridge contract is deployed")
  .addOptionalParam("quorum", "Quorum for proposals to be able to execute", 3, types.int)
  .setAction(async taskArgs => {
    relayerAddresses = JSON.parse(taskArgs.relayerAddresses);
    quorum = taskArgs.quorum;
    console.log("Relayers used for deploy", relayerAddresses);
    adminWallet = await hre.ethers.getSigner();
    console.log('Admin Public Address:', adminWallet.address);

    const ERC20Safe = await hre.ethers.getContractFactory("ERC20Safe");
    const safeContract = await ERC20Safe.deploy();
    await safeContract.deployed();
    console.log("ERC20Safe deployed to:", safeContract.address);

    const Bridge = await hre.ethers.getContractFactory("Bridge");
    const bridgeContract = await Bridge.deploy(relayerAddresses, quorum, safeContract.address);
    await bridgeContract.deployed();
    console.log("Bridge deployed to:", bridgeContract.address);
    await safeContract.setBridgeAddress(bridgeContract.address);

    fs = require('fs');
    filename = 'setup.config.json';
    data = {
      erc20Safe: safeContract.address,
      bridge: bridgeContract.address,
      relayers: relayerAddresses
    };
    fs.writeFileSync(filename, JSON.stringify(data));
  });

task("deploy-test-tokens", "Deploys ERC20 contracts to use to test the bridge")
  .setAction(async () => {
    adminWallet = await ethers.getSigner();
    const fs = require('fs');
    const filename = 'setup.config.json';
    let config = JSON.parse(fs.readFileSync(filename, 'utf8'));  
    console.log('Current contract addresses');
    console.log(config);
    const safeAddress = config["erc20Safe"];
    const safeContractFactory = await hre.ethers.getContractFactory("ERC20Safe");
    const safe = await safeContractFactory.attach(safeAddress);
    console.log("Safe at: ", safe.address);
    //deploy contracts
    const genericERC20Factory = await hre.ethers.getContractFactory("GenericERC20");

    const usdcContract = await genericERC20Factory.deploy("Dummy USDC", "dUSDC");
    await usdcContract.deployed();
    console.log("Deployed dummy USDC: ", usdcContract.address);
    const daiContract = await genericERC20Factory.deploy("Dummy DAI", "dDAI");
    await daiContract.deployed();
    console.log("Deployed dummy DAI: ", daiContract.address);
    const egldContract = await genericERC20Factory.deploy("Dummy EGLD", "dEGLD");
    await egldContract.deployed();
    console.log("Deployed dummy EGLD: ", egldContract.address);

    //whitelist tokens in safe
    console.log("Whitelisting token ", usdcContract.address);
    await safe.whitelistToken(usdcContract.address, 1);
    console.log("Whitelisting token ", daiContract.address);
    await safe.whitelistToken(daiContract.address, 1);
    console.log("Whitelisting token ", egldContract.address);
    await safe.whitelistToken(egldContract.address, 1);

    //save in configuration file
    config.tokens = [usdcContract.address, daiContract.address, egldContract.address]
    fs.writeFileSync(filename, JSON.stringify(config));
  })

task("mint-test-tokens", "Mints tests tokens and sends them to the recipientAddress")
  .addParam("recipientAddress", "Public address where the new tokens will be sent")
  .setAction(async taskArgs => {
    const recipientAddress = taskArgs.recipientAddress;
    const fs = require('fs');
    const filename = 'setup.config.json';
    let config = JSON.parse(fs.readFileSync(filename, 'utf8'));  

    for(i=0; i<config.tokens.length; i++) {
      tokenContractAddress = config.tokens[i];
      console.log('minting tokens for contract: ', tokenContractAddress);
      tokenContract = (await hre.ethers.getContractFactory("GenericERC20")).attach(tokenContractAddress);
      await tokenContract.brrr(recipientAddress);
      console.log('minted tokens for contract: ', tokenContractAddress);
    }
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
      url: "https://rinkeby.infura.io/v3/a2a8319a4e324ec4a3c5556b4c31fa08",
      accounts: {
        mnemonic: "industry layer bird test junk shadow visa lottery human spatial pact balcony"
      }
    },
    ropsten: {
      url: "https://ropsten.infura.io/v3/a2a8319a4e324ec4a3c5556b4c31fa08",
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