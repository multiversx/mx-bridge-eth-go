
const { expect } = require("chai");
const { waffle } = require("hardhat");
const { deployMockContract, provider, deployContract } = waffle;

const IERC20 = require('../artifacts/@openzeppelin/contracts/token/ERC20/IERC20.sol/IERC20.json');
const ERC20Safe = require('../artifacts/contracts/ERC20Safe.sol/ERC20Safe.json');

describe("ERC20Safe", async function () {
  const [adminWallet, otherWallet] = provider.getWallets();

  beforeEach(async function () {
    token = await deployMockContract(adminWallet, IERC20.abi);
    safe = await deployContract(adminWallet, ERC20Safe);
  });

  it('sets creator as admin', async function () {
    ADMIN_ROLE = await safe.DEFAULT_ADMIN_ROLE();
    expect(await safe.hasRole(ADMIN_ROLE, adminWallet.address)).to.be.true;
  });

  describe('whitelistToken', async function () {
    it('adds the token to the whitelistedTokens list', async function () {
      await safe.whitelistToken(token.address);

      expect(await safe._whitelistedTokens(token.address)).to.be.true;
    })
  });

  describe('deposit', async function () {
    let amount = 100;

    describe("when token is whitelisted", async function () {
      beforeEach(async function () {
        await safe.whitelistToken(token.address);
      })

      it("emits Deposited event", async () => {
        // await token.mock.transferFrom.returns({});
        await expect(safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address")))
          .to.emit(safe, "ERC20Deposited")
          .withArgs(1);
      });

      it('increments depositsCount', async () => {
        await safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address"));

        expect(await safe.depositsCount.call()).to.equal(1);
      });

      it('creates a deposit', async function () {
        await safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address"));

        deposit = await safe.getDeposit(1);

        expect(deposit.tokenAddress).to.equal(token.address);
        expect(deposit.amount).to.equal(amount);
        expect(deposit.depositor).to.equal(adminWallet.address);
        expect(deposit.status).to.equal(1/*pending*/);
        expect(ethers.utils.toUtf8String(deposit.recipient)).to.equal("some address");
      });
    });


    describe("when token is not whitelisted", async function () {
      it('reverts', async function () {
        await expect(safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address")))
          .to.be.revertedWith('Unsupported token');
      })
    });
  });

  describe('getNextPendingDeposit', async function () {
    const amount = 100;

    beforeEach(async function () {
      await safe.whitelistToken(token.address);
      await safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address"));
    });

    it('returns the deposit', async function () {
      deposit = await safe.getNextPendingDeposit();

      expect(deposit.tokenAddress).to.equal(token.address);
      expect(deposit.amount).to.equal(amount);
      expect(deposit.depositor).to.equal(adminWallet.address);
      expect(deposit.status).to.equal(1/*pending*/);
      expect(ethers.utils.toUtf8String(deposit.recipient)).to.equal("some address");
    });

    describe('when there are no pending deposits', async function() {
      beforeEach(async function() {

      });
    });
  })
});
