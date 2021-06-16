
const { expect } = require("chai");
const { waffle } = require("hardhat");
const { provider, deployContract } = waffle;

const AFC = require('../artifacts/contracts/AFCoin.sol/AFCoin.json');
const ERC20Safe = require('../artifacts/contracts/ERC20Safe.sol/ERC20Safe.json');
const Bridge = require('../artifacts/contracts/Bridge.sol/Bridge.json');
const { ethers } = require("ethers");

describe("ERC20Safe", async function () {
  const [adminWallet, bridgeWallet, otherWallet] = provider.getWallets();
  const boardMembers = [adminWallet];

  beforeEach(async function () {
    afc = await deployContract(adminWallet, AFC, [1000]);
    safe = await deployContract(adminWallet, ERC20Safe);
    bridge = await deployContract(adminWallet, Bridge, [boardMembers.map(m => m.address), 3, safe.address]);
    await afc.approve(safe.address, 1000);
    await safe.setBridgeAddress(bridge.address);
  });

  it('sets creator as admin', async function () {
    expect(await safe.adminAddress.call()).to.equal(adminWallet.address);
  });

  describe('whitelistToken', async function () {
    it('adds the token to the whitelistedTokens list', async function () {
      await safe.whitelistToken(afc.address);

      expect(await safe.whitelistedTokens(afc.address)).to.be.true;
    })

    it('emits event', async function () {
      await expect(safe.whitelistToken(afc.address))
        .to.emit(safe, 'TokenWhitelisted')
        .withArgs(afc.address);
    })

    describe('called by non admin', async function () {
      beforeEach(async function () {
        nonAdminSafe = safe.connect(otherWallet);
      });

      it('reverts', async function () {
        await (expect(nonAdminSafe.whitelistToken(afc.address))).to.be.revertedWith("Access Control: sender is not Admin");
      })
    })
  });

  describe('setBridgeAddress', async function () {
    it('updates updates the address', async function () {
      await safe.setBridgeAddress(bridgeWallet.address);

      expect(await safe.bridgeAddress.call()).to.equal(bridgeWallet.address);
    })

    it('emits event', async function () {
      await expect(safe.setBridgeAddress(bridgeWallet.address))
        .to.emit(safe, 'BridgeAddressChanged')
        .withArgs(bridgeWallet.address);
    })

    describe('called by non admin', async function () {
      beforeEach(async function () {
        nonAdminSafe = safe.connect(otherWallet);
      });

      it('reverts', async function () {
        await (expect(nonAdminSafe.setBridgeAddress(bridgeWallet.address))).to.be.revertedWith("Access Control: sender is not Admin");
      })
    })
  })

  describe('setBatchSize', async function () {
    beforeEach(async function () {
      newBatchSize = 3;
    })

    it('updates the batch size', async function () {
      await safe.setBatchSize(newBatchSize);

      expect(await safe.batchSize.call()).to.equal(newBatchSize);
    })

    it('emits event', async function () {
      await expect(safe.setBatchSize(newBatchSize))
        .to.emit(safe, 'BatchSizeChanged')
        .withArgs(newBatchSize);
    })

    describe('called by non admin', async function () {
      beforeEach(async function () {
        nonAdminSafe = safe.connect(otherWallet);
      });

      it('reverts', async function () {
        await (expect(nonAdminSafe.setBatchSize(newBatchSize))).to.be.revertedWith("Access Control: sender is not Admin");
      })
    })

    describe('called with a value above the maximum', async function () {
      it('reverts', async function () {
        await expect(safe.setBatchSize(21))
          .to.be.revertedWith("Batch size too high");
      })
    })
  })

  describe('setBatchBlockCountLimit', async function () {
    beforeEach(async function () {
      newBatchBlockCountLimit = 3;
    })

    it('updates the batch block limit', async function () {
      await safe.setBatchBlockCountLimit(newBatchBlockCountLimit);

      expect(await safe.batchBlockCountLimit.call()).to.equal(newBatchBlockCountLimit);
    })

    it('emits event', async function () {
      await expect(safe.setBatchBlockCountLimit(newBatchBlockCountLimit))
        .to.emit(safe, 'BatchBlockCountLimitChanged')
        .withArgs(newBatchBlockCountLimit);
    })

    describe('called by non admin', async function () {
      beforeEach(async function () {
        nonAdminSafe = safe.connect(otherWallet);
      });

      it('reverts', async function () {
        await (expect(nonAdminSafe.setBatchBlockCountLimit(newBatchBlockCountLimit))).to.be.revertedWith("Access Control: sender is not Admin");
      })
    })
  })

  describe('deposit', async function () {
    let amount = 100;

    describe("when token is whitelisted", async function () {
      beforeEach(async function () {
        await safe.whitelistToken(afc.address);
      })

      it("emits Deposited event", async () => {
        await expect(safe.deposit(afc.address, amount, ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq")))
          .to.emit(safe, "ERC20Deposited")
          .withArgs(1);
      });

      it('increments depositsCount', async () => {
        await safe.deposit(afc.address, amount, ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));

        expect(await safe.depositsCount.call()).to.equal(1);
      });
    });


    describe("when token is not whitelisted", async function () {
      it('reverts', async function () {
        await expect(safe.deposit(afc.address, amount, ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq")))
          .to.be.revertedWith('Unsupported token');
      })
    });
  });
});
