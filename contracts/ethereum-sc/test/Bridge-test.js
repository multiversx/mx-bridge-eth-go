const { waffle } = require("hardhat");
const { expect } = require("chai");
const { provider, deployContract } = waffle;
const { contract } = require("ethers");
const BridgeContract = require('../artifacts/contracts/Bridge.sol/Bridge.json');

describe("Bridge", async function () {
  const [adminWallet, otherWallet, relayer1, relayer2, relayer3, relayer4] = provider.getWallets();
  const boardMembers = [adminWallet, relayer1, relayer2, relayer3];
  const quorum = 3;

  beforeEach(async function () {
    bridge = await deployContract(adminWallet, BridgeContract, [boardMembers.map(m => m.address), quorum]);
  });

  it('Sets creator as admin', async function () {
    ADMIN_ROLE = await bridge.DEFAULT_ADMIN_ROLE();
    expect(await bridge.hasRole(ADMIN_ROLE, adminWallet.address)).to.be.true;
  });

  it('Sets the quorum', async function () {
    expect(await bridge._quorum.call()).to.equal(quorum);
  });

  it('Sets the board members with relayer rights', async function () {
    RELAYER_ROLE = await bridge.RELAYER_ROLE();

    boardMembers.forEach(async function (member) {
      expect(await bridge.hasRole(RELAYER_ROLE, member.address)).to.be.true;
    })
  });

  describe("addRelayer", async function () {
    it('reverts when not called by admin', async function () {
      nonAdminBridge = bridge.connect(otherWallet);
      await expect(nonAdminBridge.addRelayer(relayer4.address)).to.be.revertedWith("AccessControl: sender must be an admin to grant");
    });

    it('adds the address as a relayer', async function () {
      RELAYER_ROLE = await bridge.RELAYER_ROLE();

      await bridge.addRelayer(relayer4.address);

      expect(await bridge.hasRole(RELAYER_ROLE, relayer4.address)).to.be.true
    });

    it('emits event that a relayer was added', async function () {
      await expect(bridge.addRelayer(relayer4.address))
        .to.emit(bridge, "RelayerAdded")
        .withArgs(relayer4.address);
    });

    it('reverts if new relayer is already a relayer', async function () {
      RELAYER_ROLE = await bridge.RELAYER_ROLE();
      await bridge.addRelayer(relayer4.address);

      await expect(bridge.addRelayer(relayer4.address)).to.be.revertedWith('newRelayerAddress is already a relayer');
    });
  });

  describe('setQuorum', async function () {
    const newQuorum = 2;

    it('sets the quorum with the new value', async function () {
      await bridge.setQuorum(newQuorum);

      expect(await bridge._quorum.call()).to.equal(newQuorum);
    });

    it('reverts when not called by admin', async function () {
      nonAdminBridge = bridge.connect(otherWallet);
      await expect(nonAdminBridge.setQuorum(newQuorum)).to.be.revertedWith("Access Control: sender is not Admin");
    });
  });
});