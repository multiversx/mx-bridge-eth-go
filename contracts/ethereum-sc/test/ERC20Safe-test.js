
const { expect } = require("chai");
const { waffle } = require("hardhat");
const { deployMockContract, provider, deployContract } = waffle;

const IERC20 = require('../artifacts/@openzeppelin/contracts/token/ERC20/IERC20.sol/IERC20.json');
const ERC20Safe = require('../artifacts/contracts/ERC20Safe.sol/ERC20Safe.json');
const Bridge = require('../artifacts/contracts/Bridge.sol/Bridge.json');

describe("ERC20Safe", async function() {
  const [adminWallet, bridgeWallet, otherWallet] = provider.getWallets();

  beforeEach(async function() {
    token = await deployMockContract(adminWallet, IERC20.abi);
    bridge = await deployMockContract(adminWallet, Bridge.abi);
    safe = await deployContract(adminWallet, ERC20Safe);
  });

  it('sets creator as admin', async function() {
    ADMIN_ROLE = await safe.DEFAULT_ADMIN_ROLE();
    expect(await safe.hasRole(ADMIN_ROLE, adminWallet.address)).to.be.true;
  });

  describe('whitelistToken', async function() {
    it('adds the token to the whitelistedTokens list', async function() {
      await safe.whitelistToken(token.address);

      expect(await safe._whitelistedTokens(token.address)).to.be.true;
    })

    describe('called by non admin', async function() {
      beforeEach(async function() {
        nonAdminSafe = safe.connect(otherWallet);
      });

      it('reverts', async function() {
        await(expect(nonAdminSafe.whitelistToken(token.address))).to.be.revertedWith("Access Control: sender is not Admin");
      })
    }) 
  });

  describe('setBridgeAddress', async function() {
    it('updates updates the address', async function() {
      await safe.setBridgeAddress(bridgeWallet.address);

      expect(await safe._bridgeAddress.call()).to.equal(bridgeWallet.address);
    })

    describe('called by non admin', async function() {
      beforeEach(async function() {
        nonAdminSafe = safe.connect(otherWallet);
      });

      it('reverts', async function() {
        await(expect(nonAdminSafe.setBridgeAddress(bridgeWallet.address))).to.be.revertedWith("Access Control: sender is not Admin");
      })
    }) 
  })

  describe('deposit', async function() {
    let amount = 100;

    describe("when token is whitelisted", async function() {
      beforeEach(async function() {
        await safe.whitelistToken(token.address);
      })

      it("emits Deposited event", async () => {
        // await token.mock.transferFrom.returns({});
        await expect(safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address")))
          .to.emit(safe, "ERC20Deposited")
          .withArgs(0);
      });

      it('increments depositsCount', async () => {
        await safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address"));

        expect(await safe.depositsCount.call()).to.equal(1);
      });

      it('creates a deposit', async function() {
        await safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address"));

        deposit = await safe.getDeposit(0);

        expect(deposit.nonce).to.equal(0);
        expect(deposit.tokenAddress).to.equal(token.address);
        expect(deposit.amount).to.equal(amount);
        expect(deposit.depositor).to.equal(adminWallet.address);
        expect(deposit.status).to.equal(1/*pending*/);
        expect(ethers.utils.toUtf8String(deposit.recipient)).to.equal("some address");
      });
    });


    describe("when token is not whitelisted", async function() {
      it('reverts', async function() {
        await expect(safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address")))
          .to.be.revertedWith('Unsupported token');
      })
    });
  });

  describe('getNextPendingDeposit', async function() {
    beforeEach(async function() {
      await safe.whitelistToken(token.address);
    });
    
    describe('when there is a pending deposit', async function() {
      const amount = 100;

      beforeEach(async function() {
        await safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address"));
      });

      it('returns the deposit', async function() {
        deposit = await safe.getNextPendingDeposit();
  
        expect(deposit.nonce).to.equal(0);
        expect(deposit.tokenAddress).to.equal(token.address);
        expect(deposit.amount).to.equal(amount);
        expect(deposit.depositor).to.equal(adminWallet.address);
        expect(deposit.status).to.equal(1/*pending*/);
        expect(ethers.utils.toUtf8String(deposit.recipient)).to.equal("some address");
      });
    });
    
    describe('when there are no pending deposits', async function() {
      it('returns an empty deposit', async function() {
        deposit = await safe.getNextPendingDeposit();

        expect(deposit.nonce).to.equal(0);
        expect(deposit.tokenAddress).to.equal(ethers.constants.AddressZero);
        expect(deposit.amount).to.equal(0);
        expect(deposit.depositor).to.equal(ethers.constants.AddressZero);
        expect(deposit.status).to.equal(0/*None*/);
        expect(ethers.utils.toUtf8String(deposit.recipient)).to.equal("");
      })
    });
  });

  describe('finishCurrentPendingDeposit', async function() {
    const amount = 100;
    beforeEach(async function() {
      await safe.whitelistToken(token.address);
      await safe.setBridgeAddress(bridgeWallet.address);
      await safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address"));
      safeFromBridge = safe.connect(bridgeWallet);
    });

    it('sets the status for the deposit to Executed', async function() {
      await safeFromBridge.finishCurrentPendingDeposit(3);

      deposit = await safe.getDeposit(0);

      expect(deposit.status).to.equal(3);
    });

    describe('when there are other pending deposits', async function() {
      beforeEach(async function() {
        await safe.deposit(token.address, amount, ethers.utils.toUtf8Bytes("some address"));
      })

      it('moves to the next one', async function() {
        await safeFromBridge.finishCurrentPendingDeposit(3);
        deposit = await safe.getNextPendingDeposit();

        expect(deposit.nonce).to.equal(1);
        expect(deposit.status).to.equal(1);
      })
    })

    it('returns empty deposit if there are no other pending deposits', async function() {
      await safeFromBridge.finishCurrentPendingDeposit(3);
      deposit = await safe.getNextPendingDeposit();

      expect(deposit.nonce).to.equal(0);
      expect(deposit.status).to.equal(0);
    })

    describe('called by other than bridge', async function() {
      it('reverts', async function() {
        safeFromNonBridge = safe.connect(otherWallet);
        await(expect(safeFromNonBridge.finishCurrentPendingDeposit(3)).to.be.revertedWith("Access Control: sender is not Bridge"));
      })
    }) 
  });
});
