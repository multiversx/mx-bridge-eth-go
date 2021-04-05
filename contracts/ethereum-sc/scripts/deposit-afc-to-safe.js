// We require the Hardhat Runtime Environment explicitly here. This is optional 
// but useful for running the script in a standalone fashion through `node <script>`.
//
// When running the script with `hardhat run <script>` you'll find the Hardhat
// Runtime Environment's members available in the global scope.
const hre = require("hardhat");

async function main() {
  const tokenContractFactory = await ethers.getContractFactory("AFCoin");
  const safeContractFactory = await hre.ethers.getContractFactory("ERC20Safe");
  const tokenAddress = '0x3358F984e9B3CBBe976eEFE9B6fb92a214162932';
  const safeAddress = '0x3Aa338c8d5E6cefE95831cD0322b558677abA0f1';

  const safe = await ethers.getContractAt('ERC20Safe', safeAddress);

  await safe.depositERC20(tokenAddress, 1);

  const token = await ethers.getContractAt('AFCoin', tokenAddress);
  console.log(await token.balanceOf(safeAddress));
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main()
  .then(() => process.exit(0))
  .catch(error => {
    console.error(error);
    process.exit(1);
  });
