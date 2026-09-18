// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ERC1155} from "@openzeppelin/contracts/token/ERC1155/ERC1155.sol";
import {ERC1155Supply} from "@openzeppelin/contracts/token/ERC1155/extensions/ERC1155Supply.sol";
import {ERC2981} from "@openzeppelin/contracts/token/common/ERC2981.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {IERC165} from "@openzeppelin/contracts/utils/introspection/IERC165.sol";

/// @title Wizarding Items
/// @notice Tradable rare wands, robes, brooms, and beasts minted from the game.
/// @dev Minting happens only when the internal ledger confirms the in-game
///      item was burned. Secondary royalties follow ERC-2981 with the split
///      (treasury 5%, crafter 2.5%) handled by a splitter receiver per item.
contract WizardingItems is ERC1155, ERC1155Supply, ERC2981, AccessControl {
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");

    /// @notice Royalty in basis points paid on secondary sales (7.5%).
    uint96 public constant ROYALTY_BPS = 750;

    /// @notice Crafter of record for each item id, used for royalty splits.
    mapping(uint256 => address) public crafterOf;

    event ItemMinted(uint256 indexed id, address indexed to, address indexed crafter, bytes32 gameItemHash);

    constructor(address admin, address treasuryReceiver, string memory uri_) ERC1155(uri_) {
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _setDefaultRoyalty(treasuryReceiver, ROYALTY_BPS);
    }

    /// @notice Mint a game item that has been burned in the internal ledger.
    /// @param gameItemHash Hash of the in-game item record for audit.
    /// @param royaltyReceiver Splitter contract paying treasury and crafter.
    function mintFromGame(
        address to,
        uint256 id,
        uint256 amount,
        address crafter,
        address royaltyReceiver,
        bytes32 gameItemHash
    ) external onlyRole(MINTER_ROLE) {
        if (crafterOf[id] == address(0)) {
            crafterOf[id] = crafter;
            _setTokenRoyalty(id, royaltyReceiver, ROYALTY_BPS);
        }
        _mint(to, id, amount, "");
        emit ItemMinted(id, to, crafter, gameItemHash);
    }

    function setURI(string memory uri_) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _setURI(uri_);
    }

    function _update(address from, address to, uint256[] memory ids, uint256[] memory values)
        internal
        override(ERC1155, ERC1155Supply)
    {
        super._update(from, to, ids, values);
    }

    function supportsInterface(bytes4 interfaceId)
        public
        view
        override(ERC1155, ERC2981, AccessControl)
        returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }
}
