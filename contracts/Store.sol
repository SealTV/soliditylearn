// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

contract Store {
    event ItemSet(bytes32 key, bytes32 value); 
    mapping (bytes32 => bytes32) public items; 

    string public version;
    address payable public owner;

    constructor(string memory _version) payable {
        version = _version;
        owner = payable(msg.sender);
    }

    function setItem(bytes32 key, bytes32 value) external {
        items[key] = value;
        emit ItemSet(key, value);
    }
}