const { waffle, ethers } = require("hardhat");
const { expect } = require("chai");
const { provider, deployContract } = waffle;

const BridgeContract = require('../artifacts/contracts/Bridge.sol/Bridge.json');
const ERC20SafeContract = require('../artifacts/contracts/ERC20Safe.sol/ERC20Safe.json');
const IERC20 = require('../artifacts/@openzeppelin/contracts/token/ERC20/IERC20.sol/IERC20.json');

const { deployMockContract } = require("@ethereum-waffle/mock-contract");

describe("Bridge", async function () {
  const [adminWallet, otherWallet, relayer1, relayer2, relayer3, relayer4] = provider.getWallets();
  const boardMembers = [adminWallet, relayer1, relayer2, relayer3];
  const quorum = 3;

  beforeEach(async function () {
    mockERC20Safe = await deployMockContract(adminWallet, ERC20SafeContract.abi);
    mockERC20 = await deployMockContract(adminWallet, IERC20.abi);
    bridge = await deployContract(adminWallet, BridgeContract, [boardMembers.map(m => m.address), quorum, mockERC20Safe.address]);
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

  describe('getNextPendingTransaction', async function () {
    beforeEach(async function () {
      expectedDeposit = {
        nonce: 1,
        tokenAddress: mockERC20Safe.address,
        amount: 100,
        depositor: adminWallet.address,
        recipient: ethers.utils.toUtf8Bytes('some address'),
        status: 1
      };

      await mockERC20Safe.mock.getNextPendingDeposit.returns(expectedDeposit);
    });

    it('returns the deposit', async function () {
      transaction = await bridge.getNextPendingTransaction();


      expect(transaction['amount']).to.equal(expectedDeposit.amount);
      expect(transaction['nonce']).to.equal(expectedDeposit.nonce);
      expect(transaction['tokenAddress']).to.equal(expectedDeposit.tokenAddress);
      expect(transaction['depositor']).to.equal(expectedDeposit.depositor);
      expect(ethers.utils.toUtf8String(transaction['recipient'])).to.equal(ethers.utils.toUtf8String(expectedDeposit.recipient));
      expect(transaction['status']).to.equal(expectedDeposit.status);
    })
  });

  describe('finishCurrentPendingTransaction', async function () {
    async function mockCurrentPendingDepositDeposit() {
      depositNonce = 1;
      deposit = {
        nonce: depositNonce,
        tokenAddress: mockERC20Safe.address,
        amount: 100,
        depositor: adminWallet.address,
        recipient: ethers.utils.toUtf8Bytes('some address'),
        status: 1
      };

      await mockERC20Safe.mock.getNextPendingDeposit.returns(expectedDeposit);

      return deposit;
    }

    function getDataToSign(depositNonce, newDepositStatus) {
      signMessageDefinition = ['uint256', 'uint8', 'string'];
      signMessageData = [depositNonce, newDepositStatus, 'CurrentPendingTransaction'];

      bytesToSign = ethers.utils.solidityPack(signMessageDefinition, signMessageData);
      signData = ethers.utils.keccak256(bytesToSign);
      return ethers.utils.arrayify(signData);
    }

    async function getSignaturesForQuorum(depositNonce, newDepositStatus) {
      dataToSign = getDataToSign(depositNonce, newDepositStatus);
      signature1 = await adminWallet.signMessage(dataToSign);
      signature2 = await relayer1.signMessage(dataToSign);
      signature3 = await relayer2.signMessage(dataToSign);
      signature4 = await relayer3.signMessage(dataToSign);
      return [signature1, signature2, signature3, signature4];
    }

    beforeEach(async function () {
      expectedDeposit = await mockCurrentPendingDepositDeposit();
    });

    describe('for a different deposit than the current one', async function () {
      it('reverts', async function () {
        signatures = await getSignaturesForQuorum(2, 3)
        await (expect(bridge.finishCurrentPendingTransaction(2, 3, signatures))).to.be.revertedWith("Invalid deposit nonce");
      })
    });

    describe('for a non final state', async function () {
      it('reverts', async function () {
        await (expect(bridge.finishCurrentPendingTransaction(2, 2, []))).to.be.revertedWith("Non-final state. Can only be Executed or Rejected");
      })
    })

    describe('for a lower number of signatures than are required to achieve quorum', async function () {
      it('reverts', async function () {
        dataToSign = getDataToSign(1, 3);
        signature1 = await adminWallet.signMessage(dataToSign);

        await (expect(bridge.finishCurrentPendingTransaction(1, 3, [signature1]))).to.be.revertedWith("Not enough signatures to achieve quorum");
      })
    })

    describe('when quorum achieved', async function () {
      describe('and transaction was executed', async function () {
        beforeEach(async function () {
          expectedStatus = 3;
          signatures = await getSignaturesForQuorum(1, expectedStatus);
          await mockERC20Safe.mock.finishCurrentPendingDeposit.withArgs(3).returns();
        })

        it('updates the deposit', async function () {
          await expect(bridge.finishCurrentPendingTransaction(1, expectedStatus, signatures))
            .to.emit(bridge, 'FinishedTransaction')
            .withArgs(expectedDeposit.nonce, expectedStatus);
        })

        it('accepts geth signatures', async function () {
          gethSignatures = signatures.map(s => s.slice(0, s.length - 2) + (s.slice(-2) == '1b' ? '00' : '01'));

          await expect(bridge.finishCurrentPendingTransaction(1, expectedStatus, gethSignatures))
            .to.emit(bridge, 'FinishedTransaction')
            .withArgs(expectedDeposit.nonce, expectedStatus);
        })
      })

      describe('and transaction was rejected', async function () {
        beforeEach(async function () {
          expectedStatus = 4;
          signatures = await getSignaturesForQuorum(1, expectedStatus);
          await mockERC20Safe.mock.finishCurrentPendingDeposit.withArgs(expectedStatus).returns();
        })

        it('sets the status to reverted', async function () {
          await expect(bridge.finishCurrentPendingTransaction(1, expectedStatus, signatures))
            .to.emit(bridge, 'FinishedTransaction')
            .withArgs(expectedDeposit.nonce, expectedStatus);
        })
      })
    });
  })
});