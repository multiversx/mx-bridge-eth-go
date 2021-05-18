//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "hardhat/console.sol";
import "./SharedStructs.sol";
import "./ERC20Safe.sol";

contract Bridge is AccessControl {
    event RelayerAdded(address newRelayer);
    event FinishedTransaction(uint256 depositNonce, DepositStatus status);

    string constant action = 'CurrentPendingTransaction';
    string constant executeTransferAction = 'ExecuteTransfer';
    string constant prefix = "\x19Ethereum Signed Message:\n32";

    // Role used to execute deposits
    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");
    uint256 public _quorum;
    address private _erc20SafeAddress;
    mapping(uint256 => bool) public _executedTransfers;

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

    function addRelayer(address newRelayerAddress) external {
        require(
            !hasRole(RELAYER_ROLE, newRelayerAddress),
            "newRelayerAddress is already a relayer"
        );
        grantRole(RELAYER_ROLE, newRelayerAddress);
        emit RelayerAdded(newRelayerAddress);
    }

    function setQuorum(uint256 newQorum) external onlyAdmin {
        _quorum = newQorum;
    }

    function getNextPendingTransaction()
        external
        view
        returns (Deposit memory)
    {
        ERC20Safe safe = ERC20Safe(_erc20SafeAddress);
        return safe.getNextPendingDeposit();
    }

    function finishCurrentPendingTransaction(
        uint256 depositNonce,
        DepositStatus newDepositStatus,
        bytes[] memory signatures
    ) public {
        require(
            newDepositStatus == DepositStatus.Executed || newDepositStatus == DepositStatus.Rejected, 
            'Non-final state. Can only be Executed or Rejected');
        require(
            signatures.length >= _quorum, 
            'Not enough signatures to achieve quorum');

        ERC20Safe safe = ERC20Safe(_erc20SafeAddress);
        Deposit memory deposit = safe.getNextPendingDeposit();
        require(
            deposit.nonce == depositNonce, 
            'Invalid deposit nonce');

        uint8 signersCount;
        
        bytes32 hashedSignedData = keccak256(abi.encode(depositNonce, newDepositStatus, action));
        bytes memory prefixedSignData = abi.encodePacked(prefix, hashedSignedData);
        bytes32 hashedDepositData = keccak256(prefixedSignData);
        
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

        safe.finishCurrentPendingDeposit(newDepositStatus);
        emit FinishedTransaction(depositNonce, newDepositStatus);
    }

    function executeTransfer(address token, address recipient, uint256 amount, uint256 depositNonce, bytes[] memory signatures) public {
        require(
            signatures.length >= _quorum, 
            'Not enough signatures to achieve quorum');

            uint8 signersCount;
        
        bytes32 hashedSignedData = keccak256(abi.encode(recipient, token, amount, depositNonce, executeTransferAction));
        bytes memory prefixedSignData = abi.encodePacked(prefix, hashedSignedData);
        bytes32 hashedDepositData = keccak256(prefixedSignData);
        
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

        _executedTransfers[depositNonce] = true;

        ERC20Safe safe = ERC20Safe(_erc20SafeAddress);
        safe.transfer(token, amount, recipient);
    }

    function wasTransactionExecuted(uint256 nonceId) external view returns(bool) {
        ERC20Safe safe = ERC20Safe(_erc20SafeAddress);
        Deposit memory deposit = safe.getDeposit(nonceId);
        return deposit.status == DepositStatus.Executed || deposit.status == DepositStatus.Rejected;
    }

    function wasTransferExecuted(uint256 depositNonce) external view returns(bool) {
        return _executedTransfers[depositNonce];
    }
}
