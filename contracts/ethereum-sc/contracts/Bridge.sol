//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "hardhat/console.sol";
import "./SharedStructs.sol";
import "./ERC20Safe.sol";

/**
@title Bridge
@author Elrond & AgileFreaks
@notice Contract to be used by the bridge relayers, 
to get information and execute batches of transactions 
to be bridged.
@notice Implements access control. 
The deployer is also the admin of the contract.
In order to use it:
- relayers need to first be whitelisted
- the ERC20 safe contract must be deployed
- the safe must be setup to work in conjunction with the bridge (whitelisting)
@dev This contract mimics a multisign contract by sending the signatures from all 
relayers with the execute call, in order to save gas.
 */
contract Bridge is AccessControl {
    event RelayerAdded(address newRelayer);
    event RelayerRemoved(address removedRelayer);
    event QuorumChanged(uint256 quorum);

    string private constant action = "CurrentPendingBatch";
    string private constant executeTransferAction = "ExecuteBatchedTransfer";
    string private constant prefix = "\x19Ethereum Signed Message:\n32";
    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");

    uint256 public quorum;
    uint256 private minimumQuorum = 1;
    address private immutable erc20SafeAddress;
    mapping(uint256 => bool) public executedBatches;

    modifier onlyAdmin() {
        require(
            hasRole(DEFAULT_ADMIN_ROLE, msg.sender),
            "Access Control: sender is not Admin"
        );
        _;
    }

    modifier onlyRelayer() {
        require(
            hasRole(RELAYER_ROLE, msg.sender),
            "Access Control: sender is not Relayer"
        );
        _;
    }

    constructor(
        address[] memory board,
        uint256 intialQuorum,
        address erc20Safe
    ) {
        // whoever deploys the contract is the admin
        // DEFAULT_ADMIN_ROLE means that it can:
        //   - adjust access control
        //   - add/remove relayers
        //   - add/remove tokens that can be bridged
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);

        for (uint256 i = 0; i < board.length; i++) {
            grantRole(RELAYER_ROLE, board[i]);
        }

        require(intialQuorum >= minimumQuorum, "Quorum is too low.");
        quorum = intialQuorum;
        erc20SafeAddress = erc20Safe;
    }

    /**
        @notice Adds (whitelists) a relayer. This does not have any effect on the quorum variable.
        @param newRelayerAddress Wallet address for the new relayer that's added
    */
    function addRelayer(address newRelayerAddress) external {
        require(
            newRelayerAddress != address(0),
            "newRelayerAddress cannot be 0x0"
        );
        require(
            !hasRole(RELAYER_ROLE, newRelayerAddress),
            "newRelayerAddress is already a relayer"
        );
        grantRole(RELAYER_ROLE, newRelayerAddress);
        emit RelayerAdded(newRelayerAddress);
    }

    /**
        @notice Removes a relayer. This does not have any effect on the quorum variable.
        @param relayerAddress Wallet address for the new relayer that will be removed
    */
    function removeRelayer(address relayerAddress) external {
        require(
            hasRole(RELAYER_ROLE, relayerAddress),
            "Provided address is not a relayer"
        );
        revokeRole(RELAYER_ROLE, relayerAddress);
        emit RelayerRemoved(relayerAddress);
    }

    /**
        @notice Modifies the quorum that is needed to validate executions
        @param newQuorum Number of valid signatures required for executions. 
    */
    function setQuorum(uint256 newQuorum) external onlyAdmin {
        require(newQuorum >= minimumQuorum, "Quorum is too low.");
        quorum = newQuorum;
        emit QuorumChanged(newQuorum);
    }

    /**
        @notice Gets information about the current batch of deposits
        @return Batch which consists of:
        - batch nonce
        - timestamp
        - deposits List of the deposits included in this batch
        @dev Even if there are deposits in the Safe, the current batch might still return as empty. This is because it might not be final (not full, and not enough blocks elapsed)
    */
    function getNextPendingBatch() external view returns (Batch memory) {
        ERC20Safe safe = ERC20Safe(erc20SafeAddress);
        return safe.getNextPendingBatch();
    }

    /**
        @notice Marks all transactions from the batch with their execution status (Rejected or Executed).
        @dev This is for the Ethereum to Elrond flow
        @param batchNonceETHElrond Nonce for the batch. Should be equal to the nonce of the current batch. This identifies a batch created on the Ethereum chain toat bridges tokens from Ethereum to Elrond
        @param newDepositStatuses Array containing new statuses for all the transactions in the batch. Can only be Rejected or Executed statuses. Number of statuses must be equal to the number of transactions in the batch.
        @param signatures Signatures from all the relayers for the execution. This mimics a delegated multisig contract. For the execution to take place, there must be enough valid signatures to achieve quorum.
    */
    function finishCurrentPendingBatch(
        uint256 batchNonceETHElrond,
        DepositStatus[] calldata newDepositStatuses,
        bytes[] calldata signatures
    ) public onlyRelayer {
        for (uint256 i = 0; i < newDepositStatuses.length; i++) {
            require(
                newDepositStatuses[i] == DepositStatus.Executed ||
                    newDepositStatuses[i] == DepositStatus.Rejected,
                "Non-final state. Can only be Executed or Rejected"
            );
        }

        require(
            signatures.length >= quorum,
            "Not enough signatures to achieve quorum"
        );

        ERC20Safe safe = ERC20Safe(erc20SafeAddress);
        Batch memory batch = safe.getNextPendingBatch();
        require(batch.nonce == batchNonceETHElrond, "Invalid batch nonce");

        bytes32 hashedSignedData = keccak256(
            abi.encode(batchNonceETHElrond, newDepositStatuses, action)
        );
        bytes memory prefixedSignData = abi.encodePacked(
            prefix,
            hashedSignedData
        );
        bytes32 hashedDepositData = keccak256(prefixedSignData);
        uint256 signersCount;

        address[] memory validSigners = new address[](signatures.length);
        for (
            uint256 signatureIndex = 0;
            signatureIndex < signatures.length;
            signatureIndex++
        ) {
            bytes memory signature = signatures[signatureIndex];
            require(signature.length == 65, "Malformed signature");

            bytes32 r;
            bytes32 s;
            uint8 v;

            assembly {
                // first 32 bytes, after the length prefix
                r := mload(add(signature, 32))
                // second 32 bytes
                s := mload(add(signature, 64))
                // final byte (first byte of the next 32 bytes)
                v := byte(0, mload(add(signature, 96)))
            }

            // adjust recoverid (v) for geth cannonical values of 0 or 1
            // as per Ethereum's yellow paper: Appendinx F (Signing Transactions)
            if (v == 0 || v == 1) {
                v += 27;
            }

            address publicKey = ecrecover(hashedDepositData, v, r, s);
            require(
                hasRole(RELAYER_ROLE, publicKey),
                "Not a recognized relayer"
            );

            // Determine if we have multiple signatures from the same relayer
            uint256 si;
            for (si = 0; si < validSigners.length; si++) {
                if (validSigners[si] == address(0)) {
                    // We reached the end of the loop.
                    // This preserves the value of `si` which is used below
                    // as the first open position.
                    break;
                }

                require(
                    publicKey != validSigners[si],
                    "Multiple signatures from the same relayer"
                );
            }
            // We save this signer in the first open position.
            validSigners[si] = publicKey;
            // END: Determine if we have multiple signatures from the same relayer

            signersCount++;
        }

        require(signersCount >= quorum, "Quorum was not met");
        safe.finishCurrentPendingBatch(newDepositStatuses);
    }

    /**
        @notice Executes transfers that were signed by the relayers. 
        @dev This is for the Elrond to Ethereum flow
        @dev Arrays here try to mimmick the structure of a batch. A batch represents the values from the same index in all the arrays.
        @param tokens Array containing all the token addresses that the batch interacts with. Can even contain duplicates.
        @param recipients Array containing all the destinations from the batch. Can be duplicates.
        @param amounts Array containing all the amounts that will be transfered. 
        @param batchNonceElrondETH Nonce for the batch. This identifies a batch created on the Elrond chain that bridges tokens from Elrond to Ethereum
        @param signatures Signatures from all the relayers for the execution. This mimics a delegated multisig contract. For the execution to take place, there must be enough valid signatures to achieve quorum.
    */
    function executeTransfer(
        address[] calldata tokens,
        address[] calldata recipients,
        uint256[] calldata amounts,
        uint256 batchNonceElrondETH,
        bytes[] calldata signatures
    ) public onlyRelayer {
        require(
            signatures.length >= quorum,
            "Not enough signatures to achieve quorum"
        );
        require(
            executedBatches[batchNonceElrondETH] == false,
            "Batch already executed"
        );
        executedBatches[batchNonceElrondETH] = true;
        uint256 signersCount;

        bytes32 hashedDepositData = keccak256(
            abi.encodePacked(
                prefix,
                keccak256(
                    abi.encode(
                        recipients,
                        tokens,
                        amounts,
                        batchNonceElrondETH,
                        executeTransferAction
                    )
                )
            )
        );

        address[] memory validSigners = new address[](signatures.length);
        for (uint256 i = 0; i < signatures.length; i++) {
            bytes memory signature = signatures[i];
            require(signature.length == 65, "Malformed signature");

            bytes32 r;
            bytes32 s;
            uint8 v;

            assembly {
                // first 32 bytes, after the length prefix
                r := mload(add(signature, 32))
                // second 32 bytes
                s := mload(add(signature, 64))
                // final byte (first byte of the next 32 bytes)
                v := byte(0, mload(add(signature, 96)))
            }

            // adjust recoverid (v) for geth cannonical values of 0 or 1
            // as per Ethereum's yellow paper: Appendinx F (Signing Transactions)
            if (v == 0 || v == 1) {
                v += 27;
            }

            address publicKey = ecrecover(hashedDepositData, v, r, s);
            require(
                hasRole(RELAYER_ROLE, publicKey),
                "Not a recognized relayer"
            );

            // Determine if we have multiple signatures from the same relayer
            uint256 si;
            for (si = 0; si < validSigners.length; si++) {
                if (validSigners[si] == address(0)) {
                    // We reached the end of the loop.
                    // This preserves the value of `si` which is used below
                    // as the first open position.
                    break;
                }

                require(
                    publicKey != validSigners[si],
                    "Multiple signatures from the same relayer"
                );
            }
            // We save this signer in the first open position.
            validSigners[si] = publicKey;
            // END: Determine if we have multiple signatures from the same relayer

            signersCount++;
        }

        require(signersCount >= quorum, "Quorum was not met");

        for (uint256 j = 0; j < tokens.length; j++) {
            ERC20Safe safe = ERC20Safe(erc20SafeAddress);
            safe.transfer(tokens[j], amounts[j], recipients[j]);
        }
    }

    /**
        @notice Verifies if all the deposits within a batch are finalized (Executed or Rejected)
        @param batchNonceETHElrond Nonce for the batch.
        @return status for the batch. true - executed, false - pending (not executed yet)
    */
    function wasBatchFinished(uint256 batchNonceETHElrond)
        external
        view
        returns (bool)
    {
        ERC20Safe safe = ERC20Safe(erc20SafeAddress);
        Batch memory batch = safe.getBatch(batchNonceETHElrond);

        if (batch.deposits.length == 0) {
            return false;
        }

        for (uint256 i = 0; i < batch.deposits.length; i++) {
            if (
                batch.deposits[i].status != DepositStatus.Executed &&
                batch.deposits[i].status != DepositStatus.Rejected
            ) {
                return false;
            }
        }

        return true;
    }

    function wasBatchExecuted(uint256 batchNonceElrondETH)
        external
        view
        returns (bool)
    {
        return executedBatches[batchNonceElrondETH];
    }
}
