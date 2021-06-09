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
    event FinishedTransaction(uint256 depositNonce, DepositStatus status);
    event QuorumChanged(uint256 _quorum);

    string constant action = 'CurrentPendingBatch';
    string constant executeTransferAction = 'ExecuteBatchedTransfer';
    string constant prefix = "\x19Ethereum Signed Message:\n32";

    // Role used to execute deposits
    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");
    uint256 public _quorum;
    address private _erc20SafeAddress;
    mapping(uint256 => bool) public _executedTransfers;
    mapping(uint256 => bool) public _executedBatches;

    modifier onlyAdmin() {
        require(
            hasRole(DEFAULT_ADMIN_ROLE, msg.sender),
            "Access Control: sender is not Admin"
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

        _quorum = intialQuorum;
        _erc20SafeAddress = erc20Safe;
    }

    /**
        @notice Adds (whitelists) a relayer. This does not have any effect on the quorum variable.
        @param newRelayerAddress Wallet address for the new relayer that's added
    */
    function addRelayer(address newRelayerAddress) external {
        require(
            !hasRole(RELAYER_ROLE, newRelayerAddress),
            "newRelayerAddress is already a relayer"
        );
        grantRole(RELAYER_ROLE, newRelayerAddress);
        emit RelayerAdded(newRelayerAddress);
    }

    /**
        @notice Modifies the quorum that is needed to validate executions
        @param newQuorum Number of valid signatures required for executions. 
    */
    function setQuorum(uint256 newQuorum) external onlyAdmin {
        _quorum = newQuorum;
        emit QuorumChanged(newQuorum);
    }

    /**
        @notice Gets information about the current batch of deposits
        @return Batch which consists of:
        - batch nonce
        - startBlockNumber
        - deposits List of the deposits included in this batch
        @dev Even if there are deposits in the Safe, the current batch might still return as empty. This is because it might not be final (not full, and not enough blocks elapsed)
    */
    function getNextPendingBatch()
        external
        view
        returns (Batch memory)
    {
        ERC20Safe safe = ERC20Safe(_erc20SafeAddress);
        return safe.getNextPendingBatch();
    }

    /**
        @notice Marks all transactions from the batch with their execution status (Rejected or Executed)
        @param batchNonce Nonce for the batch. Should be equal to the nonce of the current batch.
        @param newDepositStatuses Array containing new statuses for all the transactions in the batch. Can only be Rejected or Executed statuses. Number of statuses must be equal to the number of transactions in the batch.
        @param signatures Signatures from all the relayers for the execution. This mimics a delegated multisig contract. For the execution to take place, there must be enough valid signatures to achieve quorum.
    */
    function finishCurrentPendingBatch(
        uint256 batchNonce,
        DepositStatus[] calldata newDepositStatuses,
        bytes[] calldata signatures
    ) public {
        for(uint8 i=0; i<newDepositStatuses.length; i++)
        {
            require(
                newDepositStatuses[i] == DepositStatus.Executed || newDepositStatuses[i] == DepositStatus.Rejected, 
                'Non-final state. Can only be Executed or Rejected');
        }
        
        require(
                signatures.length >= _quorum, 
                'Not enough signatures to achieve quorum');

        ERC20Safe safe = ERC20Safe(_erc20SafeAddress);
        Batch memory batch = safe.getNextPendingBatch();
        require(
                batch.nonce == batchNonce, 
                'Invalid batch nonce');
        require(
            batch.deposits.length == newDepositStatuses.length, 
            "Number of deposit statuses must match the number of deposits in the batch");

        bytes32 hashedSignedData = keccak256(abi.encode(batchNonce, newDepositStatuses, action));
        bytes memory prefixedSignData = abi.encodePacked(prefix, hashedSignedData);
        bytes32 hashedDepositData = keccak256(prefixedSignData);
        uint8 signersCount;

        for (uint256 signatureIndex = 0; signatureIndex < signatures.length; signatureIndex++) {
            bytes memory signature = signatures[signatureIndex];
            require(signature.length == 65, 'Malformed signature');

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
            if (v == 0 || v == 1)
            {
                v += 27;
            }

            address publicKey = ecrecover(hashedDepositData, v, r, s);
            require(
                hasRole(RELAYER_ROLE, publicKey),
                "Not a recognized relayer"
            );

            
            signersCount++;
        }

        require(signersCount >= _quorum, "Quorum was not met");
        safe.finishCurrentPendingBatch(newDepositStatuses);
    }

    function executeTransfer(
        address[] calldata tokens, 
        address[] calldata recipients, 
        uint256[] calldata amounts, 
        uint256 batchNonce, 
        bytes[] calldata signatures) 
    public {
        require(
            signatures.length >= _quorum, 
            'Not enough signatures to achieve quorum');

        uint8 signersCount;
        
        bytes32 hashedDepositData = keccak256(
            abi.encodePacked(
                prefix, keccak256(
                    abi.encode(
                        recipients, tokens, amounts, batchNonce, executeTransferAction))));
        
        for (uint256 i = 0; i < signatures.length; i++) {
            bytes memory signature = signatures[i];
            require(signature.length == 65, 'Malformed signature');

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
            if (v == 0 || v == 1)
            {
                v += 27;
            }

            address publicKey = ecrecover(hashedDepositData, v, r, s);
            require(
                hasRole(RELAYER_ROLE, publicKey),
                "Not a recognized relayer"
            );
            
            signersCount++;
        }

        require(signersCount >= _quorum, "Quorum was not met");

        _executedBatches[batchNonce] = true;

        for (uint8 j=0; j<tokens.length; j++)
        {
            console.log(j);
            ERC20Safe safe = ERC20Safe(_erc20SafeAddress);
            safe.transfer(tokens[j], amounts[j], recipients[j]);
        }
    }

    /**
        @notice Verifies if all the deposits within a batch are finalized (Executed or Rejected)
        @param batchNonce Nonce for the batch.
        @return status for the batch. true - executed, false - pending (not executed yet)
    */
    function wasBatchFinished(uint256 batchNonce) external view returns(bool) 
    {
        ERC20Safe safe = ERC20Safe(_erc20SafeAddress);
        Batch memory batch = safe.getBatch(batchNonce);
        
        for(uint8 i=0; i<batch.deposits.length; i++)
        {
            if(batch.deposits[i].status != DepositStatus.Executed && batch.deposits[i].status != DepositStatus.Rejected)
            {
                return false;
            }
        }
        return true;
    }

    function wasBatchExecuted(uint256 batchNonce) external view returns(bool) 
    {
        return _executedBatches[batchNonce];
    }
}
