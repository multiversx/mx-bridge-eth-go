//SPDX-License-Identifier: Unlicense
pragma solidity ^0.8.0;

enum DepositStatus {
    None, 
    Pending, 
    InProgress, //This is not used. It is here to have 1on1 mapping with statuses of deposits on the smartcontracts from Elrond 
    Executed, 
    Rejected
}
    
struct Deposit {
    uint256 nonce;
    address tokenAddress;
    uint256 amount;
    address depositor;
    bytes recipient;
    DepositStatus status;
}

struct Batch {
    uint256 nonce;
    uint timestamp;
    Deposit[] deposits;
}