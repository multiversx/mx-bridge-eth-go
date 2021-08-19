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
@notice The deployer is also the admin of the contract.
In order to use it:
- The Bridge.sol must be deployed and must be whitelisted for the Safe contract.
@dev The deposits are requested by the Bridge, and in order to save gas spent by the relayers
they will be batched either by time (batchTimeLimit) or size (batchSize).
There can only be one pending Batch. 
 */
contract ERC20Safe {
    using SafeERC20 for IERC20;
    
    uint256 public depositsCount;
    uint256 public batchesCount;
    uint256 public batchTimeLimit = 10 minutes;
    uint256 public batchSettleLimit = 10 minutes;
    // Maximum number of transactions within a batch
    uint256 public batchSize = 10;
    uint256 private constant maxBatchSize = 20;
    mapping(uint256 => Batch) public batches;
    mapping(address => bool) public whitelistedTokens;
    mapping(address => uint256) public tokenLimits;
    address public adminAddress;
    address public bridgeAddress;
    uint256 private currentPendingBatch;
    
    event BridgeAddressChanged(address newAddress);
    event BatchTimeLimitChanged(uint256 newTimeLimitInSeconds);
    event UpdatedDepositStatus(uint256 depositNonce, DepositStatus newDepositStatus);
    event BatchSizeChanged(uint256 newBatchSize);
    event TokenWhitelisted(address tokenAddress, uint256 minimumAmount);
    event TokenRemovedFromWhitelist(address tokenAddress);
    event ERC20Deposited(uint256 depositNonce);

    modifier onlyAdmin() {
        require(
            msg.sender == adminAddress, 
            "Access Control: sender is not Admin");
        _;
    }

    modifier onlyBridge() {
        require(
            msg.sender == bridgeAddress,
            "Access Control: sender is not Bridge"
        );
        _;
    }    

    constructor() {
        adminAddress = msg.sender;
    }

    /**
      @notice Whitelist a token. Only whitelisted tokens can be bridged through the bridge. 
      @param tokenAddress Address of the contract for the ERC20 token that will be used by the bridge
      @param minimumAmount Number that specifies the minimum number of tokens that the user has to deposit (this is to prevent transactions that are too small)
      @notice emits {TokenWhitelisted} event
   */
    function whitelistToken(address token, uint256 minimumAmount) external onlyAdmin {
        whitelistedTokens[token] = true;
        tokenLimits[token] = minimumAmount;
        
        emit TokenWhitelisted(token, minimumAmount);
    }

    function removeTokenFromWhitelist(address token) external onlyAdmin {
        whitelistedTokens[token] = false;
        emit TokenRemovedFromWhitelist(token);
    }

    function setBridgeAddress(address _bridgeAddress) external onlyAdmin { 
        bridgeAddress = _bridgeAddress;
        emit BridgeAddressChanged(bridgeAddress);
    }

    function setBatchTimeLimit(uint256 newBatchTimeLimit) external onlyAdmin {
        batchTimeLimit = newBatchTimeLimit;
        emit BatchTimeLimitChanged(batchTimeLimit);
    }

    function setBatchSize(uint256 newBatchSize) external onlyAdmin {
        require(newBatchSize <= maxBatchSize, "Batch size too high");
        batchSize = newBatchSize;
        emit BatchSizeChanged(batchSize);
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
        require(whitelistedTokens[tokenAddress], "Unsupported token");
        require(amount >= tokenLimits[tokenAddress], "Tried to deposit an amount below the specified limit");
        uint256 currentTimestamp = block.timestamp;

        Batch storage batch;
        if (batchesCount == 0 || batches[batchesCount-1].timestamp + batchTimeLimit < currentTimestamp || batches[batchesCount-1].deposits.length >= batchSize)
        {
            batch = batches[batchesCount];
            batch.nonce = batchesCount + 1;
            batch.timestamp = currentTimestamp;
            batchesCount++;
        }
        else
        {
            batch = batches[batchesCount - 1];
        }

        uint256 depositNonce = depositsCount+1;
        batch.deposits.push(Deposit(depositNonce, tokenAddress, amount, msg.sender, recipientAddress, DepositStatus.Pending));
        batch.lastUpdated = block.timestamp;
        depositsCount++;

        emit ERC20Deposited(depositNonce);

        IERC20 erc20 = IERC20(tokenAddress);
        erc20.safeTransferFrom(msg.sender, address(this), amount);
    }

    function transfer(address tokenAddress, uint256 amount, address recipientAddress) external onlyBridge {
        IERC20 erc20 = IERC20(tokenAddress);
        erc20.safeTransfer(recipientAddress, amount);
    }

    /**
        @notice Gets information about a batch of deposits
        @param batchNonce Identifier for the batch
        @return Batch which consists of:
        - batch nonce
        - timestamp
        - deposits List of the deposits included in this batch
    */
    function getBatch(uint256 batchNonce) 
    public
    view
    returns (Batch memory)
    {
        return batches[batchNonce-1];
    }

    /**
        @notice Gets a batch - if it is final
        @return Batch which consists of:
        - batch nonce
        - timestamp
        - deposits List of the deposits included in this batch
        @dev This function is to be called by the bridge (which is called by the relayers)
        It only returns final batches - batches where the block time limit has passed.
    */
    function getNextPendingBatch() public view returns (Batch memory) {
        Batch memory batch = batches[currentPendingBatch];

        if ((batch.lastUpdated + batchSettleLimit) < block.timestamp)
        {
            return batch;
        }
        return Batch(0, 0, 0, new Deposit[](0));
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
        Batch storage batch = batches[currentPendingBatch++];
        require(
            batch.deposits.length == statuses.length, 
            "Number of deposit statuses must match the number of deposits in the batch");
        uint256 batchDepositsCount = batch.deposits.length;
        for(uint256 i=0; i<batchDepositsCount; i++) 
        {
            batch.deposits[i].status = statuses[i];
            emit UpdatedDepositStatus(batch.deposits[i].nonce, batch.deposits[i].status);
        }
    }
}
