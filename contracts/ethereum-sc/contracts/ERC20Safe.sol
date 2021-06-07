//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "./SharedStructs.sol";
import "hardhat/console.sol";

/**
@title ERC20 Safe for bridging tokens
@author Elrond & AgileFreaks
@notice Contract to be used by the users to make deposits that will be bridged
@notice Implements access control. 
The deployer is also the admin of the contract.
In order to use it:
- The Bridge.sol must be deployed and must be whitelisted for the Safe contract.
@dev The deposits are requested by the Bridge, and in order to save gas spent by the relayers
they will be batched either by time (batchBlockCountLimit) or size (batchSize).
There can only be one pending Batch. 
 */
contract ERC20Safe is AccessControl {
    event BridgeAddressChanged(address newAddress);
    event BatchBlockCountLimitChanged(uint8 newbatchBlockCountLimit);
    event UpdatedDepositStatus(uint256 depositNonce, DepositStatus newDepositStatus);

    using SafeERC20 for IERC20;
    // STATE
    uint256 public depositsCount;
    uint256 public batchesCount;
    // Approx 10 minutes = 54 blocks * 11 sec/block = 594 sec
    uint8 public batchBlockCountLimit = 54;
    // Maximum number of transactions within a batch
    uint8 public batchSize = 10;
    mapping(uint256 => Batch) public _batches;
    
    mapping(address => bool) public _whitelistedTokens;
    address public _bridgeAddress;
    uint256 _currentPendingBatch;

    modifier onlyAdmin() {
        require(
            hasRole(DEFAULT_ADMIN_ROLE, msg.sender),
            "Access Control: sender is not Admin"
        );
        _;
    }

    modifier onlyBridge() {
        require(
            msg.sender == _bridgeAddress,
            "Access Control: sender is not Bridge"
        );
        _;
    }

    // EVENTS
    event ERC20Deposited(uint256 depositIndex);

    constructor() {
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);
    }

    function whitelistToken(address token) external onlyAdmin {
        _whitelistedTokens[token] = true;
    }

    function setBridgeAddress(address bridgeAddress) external onlyAdmin { 
        _bridgeAddress = bridgeAddress;
        emit BridgeAddressChanged(bridgeAddress);
    }

    function setBatchBlockCountLimit(uint8 newBatchBlockCountLimit) external onlyAdmin {
        batchBlockCountLimit = newBatchBlockCountLimit;
        emit BatchBlockCountLimitChanged(batchBlockCountLimit);
    }

    /**
      @notice It assumes that tokenAddress is a corect address for an ERC20 token. No checks whatsoever for this (yet)
      @param tokenAddress Address of the contract for the ERC20 token that will be deposited
      @param amount number of tokens that need to be deposited
      @param recipientAddress address of the receiver of tokens on Elrond Network
      @notice emits {ERC20Deposited} event
   */
    function deposit(
        address tokenAddress,
        uint256 amount,
        bytes calldata recipientAddress
    ) public {
        require(_whitelistedTokens[tokenAddress], "Unsupported token");
        uint256 currentBlockNumber = block.number;

        Batch storage batch;
        if (batchesCount == 0 || _batches[batchesCount-1].startBlockNumber + batchBlockCountLimit < currentBlockNumber || _batches[batchesCount-1].deposits.length >= batchSize)
        {
            batch = _batches[batchesCount];
            batch.nonce = batchesCount + 1;
            batch.startBlockNumber = currentBlockNumber;
            batchesCount++;
        }
        else
        {
            batch = _batches[batchesCount - 1];
        }

        uint256 depositIndex = depositsCount+1;
        batch.deposits.push(Deposit(depositIndex, tokenAddress, amount, msg.sender, recipientAddress, DepositStatus.Pending));
        depositsCount++;

        emit ERC20Deposited(depositIndex);
        lockTokens(tokenAddress, amount, msg.sender);
    }

    function transfer(address tokenAddress, uint256 amount, address recipientAddress) external onlyBridge {
        IERC20 erc20 = IERC20(tokenAddress);
        erc20.safeTransfer(recipientAddress, amount);
    }

    function lockTokens(
        address tokenAddress,
        uint256 amount,
        address owner
    ) internal {
        IERC20 erc20 = IERC20(tokenAddress);
        erc20.safeTransferFrom(owner, address(this), amount);
    }

    /**
        @notice Gets information about a batch of deposits
        @param batchNonce Identifier for the batch
        @return Batch which consists of:
        - batch nonce
        - startBlockNumber
        - deposits List of the deposits included in this batch
    */
    function getBatch(uint256 batchNonce) 
    public
    view
    returns (Batch memory)
    {
        return _batches[batchNonce];
    }

    /**
        @notice Gets a batch - if it is final
        @return Batch which consists of:
        - batch nonce
        - startBlockNumber
        - deposits List of the deposits included in this batch
        @dev This function is to be called by the bridge (which is called by the relayers)
        It only returns final batches - batches that are full (batchSize) or the block limit time has passed.
    */
    function getNextPendingBatch() public view returns (Batch memory) {
        Batch memory batch = _batches[_currentPendingBatch];

        if((batch.startBlockNumber + batchBlockCountLimit) < block.number || batch.deposits.length >= batchSize)
        {
            return batch;
        }
        
        return _batches[batchesCount];
    }

    /**
        @notice Marks all deposits in the current pendin batch as finalized (rejected or executed)
        @param statuses Array containing DepositStatus for each of the deposits in the current batch
        @dev This function is to be called by the bridge (which is called by the relayers)
        Updates statuses for all deposits in the batch.
        Emits event for each update
        Allows the next batch to be processed
    */
    function finishCurrentPendingBatch(DepositStatus[] calldata statuses) public onlyBridge {
        Batch storage batch = _batches[_currentPendingBatch++];

        for(uint8 i=0; i<batch.deposits.length; i++) 
        {
            batch.deposits[i].status = statuses[i];
            emit UpdatedDepositStatus(batch.deposits[i].nonce, batch.deposits[i].status);
        }
    }
}
