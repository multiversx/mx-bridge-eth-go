const { expect } = require("chai");
const { deployMockContract, MockProvider, solidity, deployContract } = require('ethereum-waffle');
const IERC20 = require('../artifacts/@openzeppelin/contracts/token/ERC20/IERC20.sol/IERC20.json');
const ERC20Safe = require('../artifacts/contracts/ERC20Safe.sol/ERC20Safe.json');
describe("ERC20Safe", function() {
  describe('depositERC20', function() {
    let token;
    let contract;
    const [wallet, otherWallet] = new MockProvider().getWallets();

    beforeEach(async function() {
      token = await deployMockContract(wallet, IERC20.abi);
      contract = await deployContract(wallet, ERC20Safe);
    });

    it("Emits Deposited event", async () => {
      let amount = 100;
      await token.mock.transferFrom.returns({});
      await expect(contract.depositERC20(token.address, amount))
        .to.emit(contract, "ERC20Deposited")
        .withArgs(token.address, wallet.address, amount)
    });
  });
});