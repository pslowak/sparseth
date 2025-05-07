// SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.20;

/**
 * @title Storage
 * @notice Store and retrieve an unsigned integer value
 */
contract Storage {
    /// @notice Emitted when a new value is stored.
    /// @param sender The address that triggered the update.
    /// @param val The value that was stored.
    event StorageUpdate(address indexed sender, uint256 val);

    /// @notice The current head of the hash chain.
    /// @dev Each time an event is emitted, the hash
    /// chain is updated. Make sure that all fields
    /// of the emitted event are used to calculate
    /// the new head.
    bytes32 private head;

    /// @notice The most recently stored value.
    uint256 private value;

    /**
     * @dev Stores 'val' in storage.
     * @param val - uint256 to be stored
     */
    function store(uint256 val) external {
        head = keccak256(abi.encode(head, msg.sender, val));
        value = val;

        emit StorageUpdate(msg.sender, val);
    }

    /**
     * @dev Returns the stored value.
     * @return the stored 'value'
     */
    function retrieve() public view returns (uint256) {
        return value;
    }
}
