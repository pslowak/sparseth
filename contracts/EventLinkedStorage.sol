// SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.20;

/**
 * @title EventLinkedStorage.sol
 * @notice Store and retrieve an unsigned integer value
 */
contract EventLinkedStorage {
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

    /// @notice A monotonic counter tracking external writes.
    /// @dev This counter is incremented every time a publicly
    /// accessible function is called that modifies state. It
    /// is used by the sparse node to verify the completeness
    /// of received transaction data involving this contract.
    ///
    /// In the context of Solidity, the counter must be
    /// incremented in functions that:
    /// - Are marked as `external` or `public`
    /// - Modify the state of the contract
    ///
    /// Functions that are marked `view` or `pure`, or those
    /// with `internal` or `private` visibility do not require
    /// the counter to be incremented, as they either do not
    /// modify state or cannot be called from outside the
    /// contract.
    uint256 private counter;

    /// @notice The most recently stored value.
    uint256 private value;

    /**
     * @dev Stores 'val' in storage.
     * @param val - uint256 to be stored
     */
    function store(uint256 val) external {
        counter++;
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
