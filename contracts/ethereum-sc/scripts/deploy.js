// We require the Hardhat Runtime Environment explicitly here. This is optional 
// but useful for running the script in a standalone fashion through `node <script>`.
//
// When running the script with `hardhat run <script>` you'll find the Hardhat
// Runtime Environment's members available in the global scope.
const hre = require("hardhat");

async function main() {
  // Hardhat always runs the compile task when running scripts with its command
  // line interface.
  //
  // If this script is run directly using `node` you may want to call compile 
  // manually to make sure everything is compiled
  // await hre.run('compile');

  // We get the contract to deploy
  [adminWallet, relayer1, relayer2, relayer3, relayer4, relayer5, depositor] = await hre.ethers.getSigners();

  console.log('Admin Public Address:', adminWallet.address);
  console.log('Relayer 1 Public Address:', relayer1.address);
  console.log('Relayer 2 Public Address:', relayer2.address);
  console.log('Relayer 3 Public Address:', relayer3.address);
  console.log('Relayer 4 Public Address:', relayer4.address);
  console.log('Relayer 5 Public Address:', relayer5.address);
  console.log('Depositor Public Address:', depositor.address);

  // Deploy ERC20 tokens
  const AFC = await hre.ethers.getContractFactory("AFCoin");
  const AFCContract = await AFC.deploy(100);
  await AFCContract.deployed();
  console.log("AFCContract deployed to:", AFCContract.address);

  depositorAFC = AFCContract.connect(depositor);
  await depositorAFC.brrr();
  console.log("Depositor created AFC: ", await depositorAFC.balanceOf(depositor.address));

  // Deploy ERC20 Safe
  const ERC20Safe = await hre.ethers.getContractFactory("ERC20Safe");
  const safeContract = await ERC20Safe.deploy();
  await safeContract.deployed();
  console.log("ERC20Safe deployed to:", safeContract.address);

  // Whitelist ERC20 tokens in the ERC20 Safe
  await safeContract.whitelistToken(AFCContract.address);

  // Deploy Bridge with ERC20 Safe address
  const Bridge = await hre.ethers.getContractFactory("Bridge");
  const bridgeContract = await Bridge.deploy([adminWallet.address, relayer1.address, relayer2.address, relayer3.address, relayer4.address, relayer5.address], 4, safeContract.address);
  await bridgeContract.deployed();
  console.log("Bridge deployed to:", bridgeContract.address);

  // Finish setup of ERC20 Safe with the Bridge so onlyBridge modifiers can be successful
  await safeContract.setBridgeAddress(bridgeContract.address);

  // Write config file
  fs = require('fs');
  filename = 'setup.config.json';
  data = {
    erc20Token: AFCContract.address,
    erc20Safe: safeContract.address,
    bridge: bridgeContract.address
  };
  fs.writeFileSync(filename, JSON.stringify(data));
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main()
  .then(() => process.exit(0))
  .catch(error => {
    console.error(error);
    process.exit(1);
  });
