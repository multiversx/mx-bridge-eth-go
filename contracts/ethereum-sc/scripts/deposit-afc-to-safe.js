// We require the Hardhat Runtime Environment explicitly here. This is optional 
// but useful for running the script in a standalone fashion through `node <script>`.
//
// When running the script with `hardhat run <script>` you'll find the Hardhat
// Runtime Environment's members available in the global scope.
const hre = require("hardhat");
const fs = require('fs');
async function main() {
  // load file
  filename = 'setup.config.json';
  config = JSON.parse(fs.readFileSync(filename, 'utf8'));
  console.log('Current contract addresses');
  console.log(config);
  // load configuration
  const tokenAddress = config["erc20Token"];
  const safeAddress = config["erc20Safe"];
  const bridgeAddress = config["bridge"];
  [adminWallet, depositor, relayer1] = await hre.ethers.getSigners();
  // load deployed contracts
  const tokenContractFactory = await hre.ethers.getContractFactory("AFCoin");
  const safeContractFactory = await hre.ethers.getContractFactory("ERC20Safe");
  const bridgeContractFactory = await hre.ethers.getContractFactory("Bridge");
  const token = await tokenContractFactory.attach(tokenAddress).connect(depositor);
  const safe = await safeContractFactory.attach(safeAddress).connect(depositor);
  const bridge = await bridgeContractFactory.attach(bridgeAddress);
  // transactions
  await token.approve(safe.address, 4);
  console.log("Balance for depositor", (await token.balanceOf(depositor.address)).toString());
  console.log("Allowance for depositor", (await token.allowance(depositor.address, safeAddress)).toString());
  await safe.deposit(token.address, 3, hre.ethers.utils.toUtf8Bytes("erd1rve9evhhfhuw26ctgctzxmevptj43yv800g9603l8vmua2ew7lcq4tp2an")); // Maria
  console.log("Balance for depositor", (await token.balanceOf(depositor.address)).toString());
  console.log("Balance in safe", (await token.balanceOf(safe.address)).toString());
  pendingBatch = await bridge.getNextPendingBatch();
  console.log(pendingBatch);
}
// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main()
  .then(() => process.exit(0))
  .catch(error => {
    console.error(error);
    process.exit(1);
  });