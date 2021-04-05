//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

// TODO: Rename to ERC20Bridge
contract ERC20Safe {
    event ERC20Deposited(
        address tokenAddress,
        address depositor,
        uint256 amount
    );

  // TODO: Add a set of whitelisted ERC20 addresses
  /**

      @notice It assumes that tokenAddress is a corect address for an ERC20 token. No checks whatsoever for this (yet)
      @param tokenAddress Address of the contract for the ERC20 token that will be deposited
      @param amount number of tokens that need to be deposited
      @notice emits {ERC20Deposited} event
   */
    function depositERC20(address tokenAddress, uint256 amount, string calldata data) public {
        IERC20 erc20 = IERC20(tokenAddress);
        // _safeTransferFrom(erc20, msg.sender, address(this), amount);
        emit ERC20Deposited(tokenAddress, msg.sender, amount);
    }

    function _safeTransfer(IERC20 token, address to, uint256 value) private {
        _safeCall(token, abi.encodeWithSelector(token.transfer.selector, to, value));
    }
    
    function _safeTransferFrom(IERC20 token, address from, address to, uint256 value) private {
        _safeCall(token, abi.encodeWithSelector(token.transferFrom.selector, from, to, value));
    }

    function _safeCall(IERC20 token, bytes memory data) private {        
        (bool success, bytes memory returndata) = address(token).delegatecall(data);
        require(success, "ERC20: call failed");

        if (returndata.length > 0) {
            require(abi.decode(returndata, (bool)), "ERC20: operation did not succeed");
        }
    }
}