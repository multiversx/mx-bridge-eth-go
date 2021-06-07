
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
    bridge = await deployContract(adminWallet, Bridge, [boardMembers.map(m => m.address), 1, safe.address]);
    await afc.approve(safe.address, 1000);
    await safe.setBridgeAddress(bridge.address);
  });

  it('sets creator as admin', async function () {
    ADMIN_ROLE = await safe.DEFAULT_ADMIN_ROLE();
    expect(await safe.hasRole(ADMIN_ROLE, adminWallet.address)).to.be.true;
  });

  describe('whitelistToken', async function () {
    it('adds the token to the whitelistedTokens list', async function () {
      await safe.whitelistToken(afc.address);

      expect(await safe._whitelistedTokens(afc.address)).to.be.true;
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

      expect(await safe._bridgeAddress.call()).to.equal(bridgeWallet.address);
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
