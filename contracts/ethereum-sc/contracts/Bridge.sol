//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "hardhat/console.sol";
import "./SharedStructs.sol";
import "./ERC20Safe.sol";

contract Bridge is AccessControl {
    event RelayerAdded(address newRelayer);

    // Role used to execute deposits
    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");
    uint256 public _quorum;
    address private _erc20SafeAddress;

    modifier onlyAdmin() {
        require(
            hasRole(DEFAULT_ADMIN_ROLE, msg.sender),
            "Access Control: sender is not Admin"
        );
        _;
    }

    constructor(address[] memory board, uint256 intialQuorum, address erc20Safe) {
        // whoever deploys the contract is the admin
        // DEFAULT_ADMIN_ROLE means that it can:
        //   - adjust access control
        //   - add/remove relayers
        //   - add/remove tokens that can be bridged
        _setupRole(DEFAULT_ADMIN_ROLE, msg.sender);

        for (uint256 i; i < board.length; i++) {
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

    function getNextPendingTransaction() external view returns(Deposit memory) {
        ERC20Safe safe = ERC20Safe(_erc20SafeAddress);
        return safe.getNextPendingDeposit();
    }

    //get next pending transaction

    //execute transaction

    //finalize pending transaction (success/fail)
}
