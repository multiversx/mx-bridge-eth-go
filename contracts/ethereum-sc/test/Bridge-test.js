const { waffle, ethers } = require("hardhat");
const { expect } = require("chai");
const { provider, deployContract } = waffle;

const BridgeContract = require('../artifacts/contracts/Bridge.sol/Bridge.json');
const ERC20SafeContract = require('../artifacts/contracts/ERC20Safe.sol/ERC20Safe.json');
const IERC20 = require('../artifacts/@openzeppelin/contracts/token/ERC20/IERC20.sol/IERC20.json');
const AFC = require('../artifacts/contracts/AFCoin.sol/AFCoin.json');

const { deployMockContract } = require("@ethereum-waffle/mock-contract");
const { toUtf8String } = require("@ethersproject/strings");

describe("Bridge", async function () {
  const [adminWallet, relayer1, relayer2, relayer3, relayer4, relayer5, relayer6, relayer7, relayer8, otherWallet] = provider.getWallets();
  const boardMembers = [adminWallet, relayer1, relayer2, relayer3, relayer5, relayer6, relayer7, relayer8];
  const quorum = 7;

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

    it('emits event', async function () {
      await expect(bridge.setQuorum(newQuorum))
        .to.emit(bridge, 'QuorumChanged')
        .withArgs(newQuorum);
    })

    it('reverts when not called by admin', async function () {
      nonAdminBridge = bridge.connect(otherWallet);
      await expect(nonAdminBridge.setQuorum(newQuorum)).to.be.revertedWith("Access Control: sender is not Admin");
    });
  });

  async function setupContracts() {
    afc = await deployContract(adminWallet, AFC, [1000]);
    erc20Safe = await deployContract(adminWallet, ERC20SafeContract);
    bridge = await deployContract(adminWallet, BridgeContract, [boardMembers.map(m => m.address), quorum, erc20Safe.address]);
    await erc20Safe.setBridgeAddress(bridge.address);
    await afc.approve(erc20Safe.address, 200);
    await erc20Safe.whitelistToken(afc.address);
    batchSize = 10;
  }

  async function setupFullBatch() {
    for (i = 0; i < batchSize; i++) {
      await erc20Safe.deposit(afc.address, 2, hre.ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));
    }
  }

  async function setupReadyBatch() {
    blockCountLimit = 3;
    await erc20Safe.setBatchBlockCountLimit(blockCountLimit);
    await erc20Safe.deposit(afc.address, 2, hre.ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));
    for (i = 0; i < blockCountLimit + 1; i++) {
      await network.provider.send("evm_mine")
    }
  }

  describe('getNextPendingBatch', async function () {
    beforeEach(async function () {
      await setupContracts();
    });

    describe('when batch is ready', async function () {
      describe('by being full', async function () {
        beforeEach(async function () {
          await setupFullBatch();
        });

        it('returns the batch', async function () {
          batch = await bridge.getNextPendingBatch();
          expect(batch.nonce).to.equal(1);
        })

        it('returns all the deposits', async function () {
          batch = await bridge.getNextPendingBatch();
          expect(batch.deposits.length).to.equal(10);
        })
      })

      describe('by being old', async function () {
        beforeEach(async function () {
          await setupReadyBatch();
        })

        it('returns the batch', async function () {
          batch = await bridge.getNextPendingBatch();
          expect(batch.nonce).to.equal(1);
        })

        it('returns all the deposits', async function () {
          batch = await bridge.getNextPendingBatch();
          expect(batch.deposits.length).to.equal(1);
        })
      })
    });

    describe('when batch is not ready', async function () {
      beforeEach(async function () {
        await erc20Safe.deposit(afc.address, 2, hre.ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));
      });

      it('returns an empty batch', async function () {
        batch = await bridge.getNextPendingBatch();
        expect(batch.nonce).to.equal(0);
      })
    })
  });

  describe('finishCurrentPendingBatch', async function () {
    async function getBatchDataToSign(batch, newStatuses) {
      signMessageDefinition = ['uint256', 'uint8[]', 'string'];
      signMessageData = [batch.nonce, newStatuses, 'CurrentPendingBatch'];
      bytesToSign = ethers.utils.defaultAbiCoder.encode(signMessageDefinition, signMessageData);
      signData = ethers.utils.keccak256(bytesToSign);

      return ethers.utils.arrayify(signData);
    }

    async function getBatchSignaturesForQuorum(batch, newStatuses) {
      dataToSign = await getBatchDataToSign(batch, newStatuses);
      signature1 = await adminWallet.signMessage(dataToSign);
      signature2 = await relayer1.signMessage(dataToSign);
      signature3 = await relayer2.signMessage(dataToSign);
      signature4 = await relayer3.signMessage(dataToSign);
      signature5 = await relayer5.signMessage(dataToSign);
      signature6 = await relayer6.signMessage(dataToSign);
      signature7 = await relayer7.signMessage(dataToSign);

      return [signature1, signature2, signature3, signature4, signature5, signature6, signature7];
    }
    beforeEach(async function () {
      await setupContracts();
      await setupFullBatch();
    });

    describe('when quorum achieved', async function () {
      describe('all deposits executed successful', async function () {
        beforeEach(async function () {
          newDepositStatuses = [3, 3, 3, 3, 3, 3, 4, 4, 4, 4];
          batch = await bridge.getNextPendingBatch();
          signatures = await getBatchSignaturesForQuorum(batch, newDepositStatuses)
        })

        it('updates the deposits', async function () {
          await expect(bridge.finishCurrentPendingBatch(batch.nonce, newDepositStatuses, signatures))
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[0].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[1].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[2].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[3].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[4].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[5].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[6].nonce, 4)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[7].nonce, 4)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[8].nonce, 4)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[9].nonce, 4);
        })

        it('accepts geth signatures', async function () {
          gethSignatures = signatures.map(s => s.slice(0, s.length - 2) + (s.slice(-2) == '1b' ? '00' : '01'));

          await expect(bridge.finishCurrentPendingBatch(batch.nonce, newDepositStatuses, gethSignatures))
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[0].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[1].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[2].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[3].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[4].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[5].nonce, 3)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[6].nonce, 4)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[7].nonce, 4)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[8].nonce, 4)
            .to.emit(erc20Safe, 'UpdatedDepositStatus')
            .withArgs(batch.deposits[9].nonce, 4);
        })

        it('moves to the next batch', async function () {
          await bridge.finishCurrentPendingBatch(batch.nonce, newDepositStatuses, signatures);

          nextBatch = await bridge.getNextPendingBatch();
          expect(nextBatch.nonce).to.not.equal(batch.nonce);
        })

        it('returns that the batch was executed', async function () {
          await bridge.finishCurrentPendingBatch(batch.nonce, newDepositStatuses, signatures);

          expect(await bridge.wasBatchExecuted(batch.nonce)).to.be.true;
        })
      })
    })

    describe('with incorrect number of statuses', async function () {
      beforeEach(async function () {
        newDepositStatuses = [3, 3];
        batch = await bridge.getNextPendingBatch();
      })
      it('reverts', async function () {
        await expect(bridge.finishCurrentPendingBatch(batch.nonce, newDepositStatuses, await getBatchSignaturesForQuorum(batch, newDepositStatuses)))
          .to.be.revertedWith("Number of deposit statuses must match the number of deposits in the batch");
      })
    })

    describe('with non final states', async function () {
      beforeEach(async function () {
        newDepositStatuses = [1, 3, 3, 3, 3, 3, 3, 3, 3, 3];
        batch = await bridge.getNextPendingBatch();
      })
      it('reverts', async function () {
        await expect(bridge.finishCurrentPendingBatch(batch.nonce, newDepositStatuses, await getBatchSignaturesForQuorum(batch, newDepositStatuses)))
          .to.be.revertedWith("Non-final state. Can only be Executed or Rejected");
      })
    })

    describe('with not enough signatures', async function () {
      beforeEach(async function () {
        newDepositStatuses = [3, 3, 3, 3, 3, 3, 3, 3, 3, 3];
        batch = await bridge.getNextPendingBatch();
      })
      it('reverts', async function () {
        signature1 = await adminWallet.signMessage('dataToSign');
        await expect(bridge.finishCurrentPendingBatch(batch.nonce, newDepositStatuses, [signature1]))
          .to.be.revertedWith("Not enough signatures to achieve quorum");
      })
    })
  })

  describe('executeTransfer', async function () {
    beforeEach(async function () {
      afc = await deployContract(adminWallet, AFC, [1000]);
      erc20Safe = await deployContract(adminWallet, ERC20SafeContract);
      bridge = await deployContract(adminWallet, BridgeContract, [boardMembers.map(m => m.address), quorum, erc20Safe.address]);
      await erc20Safe.setBridgeAddress(bridge.address);

      await afc.approve(erc20Safe.address, 200);
      await erc20Safe.whitelistToken(afc.address);
      await erc20Safe.deposit(afc.address, 200, hre.ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));
    });

    function getExecuteTransferData(tokenAddress, recipientAddress, amount) {
      let depositNonce = 42;
      signMessageDefinition = ['address', 'address', 'uint256', 'uint256', 'string'];
      signMessageData = [recipientAddress, tokenAddress, amount, depositNonce, 'ExecuteTransfer'];

      bytesToSign = ethers.utils.defaultAbiCoder.encode(signMessageDefinition, signMessageData);
      signData = ethers.utils.keccak256(bytesToSign);
      return ethers.utils.arrayify(signData);
    }

    async function getSignaturesForExecuteTransfer(tokenAddress, recipientAddress, amount) {
      dataToSign = getExecuteTransferData(tokenAddress, recipientAddress, amount);
      signature1 = await adminWallet.signMessage(dataToSign);
      signature2 = await relayer1.signMessage(dataToSign);
      signature3 = await relayer2.signMessage(dataToSign);
      signature4 = await relayer3.signMessage(dataToSign);
      signature5 = await relayer5.signMessage(dataToSign);
      signature6 = await relayer6.signMessage(dataToSign);
      signature7 = await relayer7.signMessage(dataToSign);

      return [signature1, signature2, signature3, signature4, signature5, signature6, signature7];
    }

    it('transfers tokens', async function () {
      amount = 200;
      depositNonce = 42;
      signatures = await getSignaturesForExecuteTransfer(afc.address, otherWallet.address, amount);

      await expect(() => bridge.executeTransfer(afc.address, otherWallet.address, amount, depositNonce, signatures))
        .to.changeTokenBalance(afc, otherWallet, amount);
    })

    it('sets the wasTransferExecuted to true', async function () {
      amount = 200;
      depositNonce = 42;
      signatures = await getSignaturesForExecuteTransfer(afc.address, otherWallet.address, amount);
      await bridge.executeTransfer(afc.address, otherWallet.address, amount, depositNonce, signatures);
      expect(await bridge.wasTransferExecuted(depositNonce)).to.be.true;
    })

    describe('not enough signatures for quorum', async function () {
      it('reverts', async function () {
        amount = 200;
        depositNonce = 42;
        signatures = (await getSignaturesForExecuteTransfer(afc.address, otherWallet.address, 200)).slice(0, -2);

        await expect(bridge.executeTransfer(afc.address, otherWallet.address, amount, depositNonce, signatures)).to.be.revertedWith("Not enough signatures to achieve quorum");
      })

      it('does not set wasTransferExecuted', async function () {
        expect(await bridge.wasTransferExecuted(depositNonce)).to.be.false;
      })
    })
  })
});