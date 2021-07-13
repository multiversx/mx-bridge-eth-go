const { waffle, ethers, network } = require("hardhat");
const { expect } = require("chai");
const { provider, deployContract } = waffle;

const BridgeContract = require('../artifacts/contracts/Bridge.sol/Bridge.json');
const ERC20SafeContract = require('../artifacts/contracts/ERC20Safe.sol/ERC20Safe.json');
const IERC20 = require('../artifacts/@openzeppelin/contracts/token/ERC20/IERC20.sol/IERC20.json');
const AFC = require('../artifacts/contracts/AFCoin.sol/AFCoin.json');

describe("Bridge", async function () {
  const [adminWallet, relayer1, relayer2, relayer3, relayer4, relayer5, relayer6, relayer7, relayer8, otherWallet] = provider.getWallets();
  const boardMembers = [adminWallet, relayer1, relayer2, relayer3, relayer5, relayer6, relayer7, relayer8];
  const quorum = 7;
  const batchSize = 10;

  async function setupContracts() {
    erc20Safe = await deployContract(adminWallet, ERC20SafeContract);
    bridge = await deployContract(adminWallet, BridgeContract, [boardMembers.map(m => m.address), quorum, erc20Safe.address]);
    await erc20Safe.setBridgeAddress(bridge.address);
    await setupErc20Token()
  }

  async function setupErc20Token() {
    afc = await deployContract(adminWallet, AFC, [1000]);
    await afc.approve(erc20Safe.address, 1000);
    await erc20Safe.whitelistToken(afc.address);
  }

  async function setupFullBatch() {
    for (i = 0; i < batchSize; i++) {
      await erc20Safe.deposit(afc.address, 2, hre.ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));
    }
  }

  async function settleCurrentBatch() {
    // leave enough time to consider the batch settled (probability for a reorg is minimal)
    // 1 minute and one second into the future
    settleTime = (1 * 60) + 1;
    await network.provider.send('evm_increaseTime', [settleTime]);
    await network.provider.send("evm_mine")
  }

  async function setupReadyBatch() {
    await erc20Safe.deposit(afc.address, 2, hre.ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));

    // 10 minutes and one second into the future
    timeElapsedSinceBatchCreation = (10 * 60) + 1;
    await network.provider.send('evm_increaseTime', [timeElapsedSinceBatchCreation]);
    await network.provider.send("evm_mine")
  }

  beforeEach(async function () {
    await setupContracts();
  });

  it('Sets creator as admin', async function () {
    ADMIN_ROLE = await bridge.DEFAULT_ADMIN_ROLE();
    expect(await bridge.hasRole(ADMIN_ROLE, adminWallet.address)).to.be.true;
  });

  it('Sets the quorum', async function () {
    expect(await bridge.quorum.call()).to.equal(quorum);
  });

  it('Sets the board members with relayer rights', async function () {
    RELAYER_ROLE = await bridge.RELAYER_ROLE();

    boardMembers.forEach(async function (member) {
      expect(await bridge.hasRole(RELAYER_ROLE, member.address)).to.be.true;
    })
  });

  describe('when initialized with a quorum that is lower than the minimum', async function () {
    it('reverts', async function () {
      invalidQuorumValue = 1;
      await expect(deployContract(adminWallet, BridgeContract, [boardMembers.map(m => m.address), invalidQuorumValue, erc20Safe.address]))
        .to.be.revertedWith("Quorum is too low.");
    })
  })

  describe("addRelayer", async function () {
    it('reverts when called with an empty address', async function () {
      await expect(bridge.addRelayer(ethers.constants.AddressZero)).to.be.revertedWith('newRelayerAddress cannot be 0x0');
    })

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

  describe('removeRelayer', async function () {
    beforeEach(async function () {
      RELAYER_ROLE = await bridge.RELAYER_ROLE();
      await bridge.addRelayer(relayer4.address);
    })

    it('removes the relayer', async function () {
      await bridge.removeRelayer(relayer4.address);

      expect(await bridge.hasRole(RELAYER_ROLE, relayer4.address)).to.be.false
    })

    it('emits an event', async function () {
      await expect(bridge.removeRelayer(relayer4.address))
        .to.emit(bridge, "RelayerRemoved")
        .withArgs(relayer4.address);
    })

    it('reverts when not called by admin', async function () {
      nonAdminBridge = bridge.connect(otherWallet);
      await expect(nonAdminBridge.removeRelayer(relayer4.address)).to.be.revertedWith("AccessControl: sender must be an admin to revoke");
    });

    it('reverts if address is not already a relayer', async function () {
      await expect(bridge.removeRelayer(otherWallet.address)).to.be.revertedWith('Provided address is not a relayer');
    });
  })

  describe('setQuorum', async function () {
    const newQuorum = 8;

    it('sets the quorum with the new value', async function () {
      await bridge.setQuorum(newQuorum);

      expect(await bridge.quorum.call()).to.equal(newQuorum);
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

    describe('when quorum is lower than the minimum', async function () {
      it('reverts', async function () {
        await expect(bridge.setQuorum(2)).to.be.revertedWith('Quorum is too low.');
      })
    })
  });

  describe('getNextPendingBatch', async function () {

    describe('when batch is ready', async function () {
      describe('by being full', async function () {
        beforeEach(async function () {
          await setupFullBatch();
          await settleCurrentBatch();
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
      describe('because it is not full', async function () {
        beforeEach(async function () {
          await erc20Safe.deposit(afc.address, 2, hre.ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));
        });

        it('returns an empty batch', async function () {
          batch = await bridge.getNextPendingBatch();
          expect(batch.nonce).to.equal(0);
        })
      })

      describe('because not enough time has passed since the batch was created', async function () {
        beforeEach(async function () {
          await erc20Safe.deposit(afc.address, 2, hre.ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));

          // 10 minutes into the future
          timeElapsedSinceBatchCreation = (10 * 60);
          await network.provider.send('evm_increaseTime', [timeElapsedSinceBatchCreation]);
          await network.provider.send("evm_mine")
        });

        it('returns an empty batch', async function () {
          batch = await bridge.getNextPendingBatch();
          expect(batch.nonce).to.equal(0);
        })
      })

      describe('because not enough time has passed since the last transaction in batch', async function () {
        beforeEach(async function () {
          await setupFullBatch();
        })

        it('returns an empty batch', async function () {
          batch = await bridge.getNextPendingBatch();
          expect(batch.nonce).to.equal(0);
        })
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
      await setupFullBatch();
      await settleCurrentBatch();
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

        it('returns that the batch was finsihed', async function () {
          await bridge.finishCurrentPendingBatch(batch.nonce, newDepositStatuses, signatures);

          expect(await bridge.wasBatchFinished(batch.nonce)).to.be.true;
        })
      })

      describe('but all signatures are from the same relayer', async function () {
        beforeEach(async function () {
          newDepositStatuses = [3, 3, 3, 3, 3, 3, 4, 4, 4, 4];
          batch = await bridge.getNextPendingBatch();

          dataToSign = await getBatchDataToSign(batch, newDepositStatuses);
          signature1 = await adminWallet.signMessage(dataToSign);
          signatures = [signature1, signature1, signature1, signature1, signature1, signature1, signature1];
        })

        it('reverts', async function () {
          await expect(bridge.finishCurrentPendingBatch(batch.nonce, newDepositStatuses, signatures))
            .to.be.revertedWith("Multiple signatures from the same relayer");
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

    describe('called by a non relayer', async function () {
      beforeEach(async function () {
        newDepositStatuses = [3, 3, 3, 3, 3, 3, 3, 3, 3, 3];
        batch = await bridge.getNextPendingBatch();
        signatures = await getBatchSignaturesForQuorum(batch, newDepositStatuses)
      })

      it('reverts', async function () {
        nonAdminBridge = bridge.connect(otherWallet);
        await expect(nonAdminBridge.finishCurrentPendingBatch(batch.nonce, newDepositStatuses, signatures)).to.be.revertedWith("Access Control: sender is not Relayer");
      })
    })
  })

  describe('executeTransfer', async function () {
    beforeEach(async function () {
      await erc20Safe.deposit(afc.address, 200, hre.ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));
      amount = 200;
      batchNonce = 42;
      signatures = await getSignaturesForExecuteTransfer([afc.address], [otherWallet.address], [amount], batchNonce);
    });

    function getExecuteTransferData(tokenAddresses, recipientAddresses, amounts, batchNonce) {
      signMessageDefinition = ['address[]', 'address[]', 'uint256[]', 'uint256', 'string'];
      signMessageData = [recipientAddresses, tokenAddresses, amounts, batchNonce, 'ExecuteBatchedTransfer'];

      bytesToSign = ethers.utils.defaultAbiCoder.encode(signMessageDefinition, signMessageData);
      signData = ethers.utils.keccak256(bytesToSign);
      return ethers.utils.arrayify(signData);
    }

    async function getSignaturesForExecuteTransfer(tokenAddresses, recipientAddresses, amounts, batchNonce) {
      dataToSign = getExecuteTransferData(tokenAddresses, recipientAddresses, amounts, batchNonce);
      signature1 = await adminWallet.signMessage(dataToSign);
      signature2 = await relayer1.signMessage(dataToSign);
      signature3 = await relayer2.signMessage(dataToSign);
      signature4 = await relayer3.signMessage(dataToSign);
      signature5 = await relayer5.signMessage(dataToSign);
      signature6 = await relayer6.signMessage(dataToSign);
      signature7 = await relayer7.signMessage(dataToSign);

      return [signature1, signature2, signature3, signature4, signature5, signature6, signature7];
    }

    describe("when quorum achieved", async function () {
      it('transfers tokens', async function () {
        await expect(() => bridge.executeTransfer([afc.address], [otherWallet.address], [amount], batchNonce, signatures))
          .to.changeTokenBalance(afc, otherWallet, amount);
      })

      it('sets the wasBatchExecuted to true', async function () {
        await bridge.executeTransfer([afc.address], [otherWallet.address], [amount], batchNonce, signatures);
        expect(await bridge.wasBatchExecuted(batchNonce)).to.be.true;
      })

      describe('but all signatures are from the same relayer', async function () {
        beforeEach(async function () {
          dataToSign = await getExecuteTransferData([afc.address], [otherWallet.address], [amount], batchNonce);
          signature1 = await adminWallet.signMessage(dataToSign);
          signatures = [signature1, signature1, signature1, signature1, signature1, signature1, signature1];
        })

        it('reverts', async function () {
          await expect(bridge.executeTransfer([afc.address], [otherWallet.address], [amount], batchNonce, signatures))
            .to.be.revertedWith("Multiple signatures from the same relayer");
        })
      })
    })

    describe('not enough signatures for quorum', async function () {
      it('reverts', async function () {
        await expect(bridge.executeTransfer([afc.address], [otherWallet.address], [amount], batchNonce, signatures.slice(0, -2))).to.be.revertedWith("Not enough signatures to achieve quorum");
      })

      it('does not set wasBatchExecuted', async function () {
        expect(await bridge.wasBatchExecuted(batchNonce)).to.be.false;
      })
    })

    describe('trying to replay the batch', async function () {
      beforeEach(async function () {
        // add more funds in order to not fail because of insufficient balance
        await erc20Safe.deposit(afc.address, 200, hre.ethers.utils.toUtf8Bytes("erd13kgks9km5ky8vj2dfty79v769ej433k5xmyhzunk7fv4pndh7z2s8depqq"));

        await bridge.executeTransfer([afc.address], [otherWallet.address], [amount], batchNonce, signatures)
      })

      it('reverts', async function () {
        await expect(bridge.executeTransfer([afc.address], [otherWallet.address], [amount], batchNonce, signatures)).to.be.revertedWith("Batch already executed");
      })
    })

    describe('called by a non relayer', async function () {
      it('reverts', async function () {
        nonAdminBridge = bridge.connect(otherWallet);
        await expect(nonAdminBridge.executeTransfer([afc.address], [otherWallet.address], [amount], batchNonce, signatures)).to.be.revertedWith("Access Control: sender is not Relayer");
      })
    })
  })

  describe('wasBatchFinished', async function () {
    it('is false for non-existent batch', async function () {
      expect(await bridge.wasBatchFinished(42)).to.be.false;
    })
  })
});