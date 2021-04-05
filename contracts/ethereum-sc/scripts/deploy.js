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
  const AFC = await hre.ethers.getContractFactory("AFCoin");
  const AFCContract = await AFC.deploy(10);

  await AFCContract.deployed();

  console.log("AFCContract deployed to:", AFCContract.address);

  const ERC20Safe = await hre.ethers.getContractFactory("ERC20Safe");
  const safeContract = await ERC20Safe.deploy();

  await safeContract.deployed();

  console.log("ERC20Safe deployed to:", safeContract.address);
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main()
  .then(() => process.exit(0))
  .catch(error => {
    console.error(error);
    process.exit(1);
  });
