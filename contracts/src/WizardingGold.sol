// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC20Burnable} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import {ERC20Pausable} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Pausable.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";

/// @title Wizarding Gold (WGLD)
/// @notice Tradable premium token for Hogwarts Online.
/// @dev The internal ledger (services/ledger) is the source of truth. This
///      contract mirrors settled balances. Minting is restricted to the
///      settlement role and bounded by a per-season cap that matches the
///      ledger's configured cap so the chain cannot exceed the game economy.
contract WizardingGold is ERC20, ERC20Burnable, ERC20Pausable, AccessControl {
    bytes32 public constant SETTLER_ROLE = keccak256("SETTLER_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");

    /// @notice Hard cap on total supply ever minted (genesis plus all seasons).
    uint256 public immutable maxSupply;

    /// @notice Emission cap per season id.
    mapping(uint256 => uint256) public seasonCap;
    /// @notice Emission minted so far per season id.
    mapping(uint256 => uint256) public seasonMinted;

    event SeasonCapSet(uint256 indexed season, uint256 cap);
    event SeasonMint(uint256 indexed season, address indexed to, uint256 amount);

    error SeasonCapExceeded(uint256 season, uint256 requested, uint256 remaining);
    error MaxSupplyExceeded(uint256 requested, uint256 remaining);

    constructor(address admin, address treasury, uint256 genesisSupply, uint256 maxSupply_)
        ERC20("Wizarding Gold", "WGLD")
    {
        require(genesisSupply <= maxSupply_, "genesis exceeds max");
        maxSupply = maxSupply_;
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(PAUSER_ROLE, admin);
        _mint(treasury, genesisSupply);
    }

    /// @notice Set the emission cap for a season. Can only be lowered once set
    ///         above zero, so governance cannot quietly inflate mid-season.
    function setSeasonCap(uint256 season, uint256 cap) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(seasonCap[season] == 0 || cap < seasonCap[season], "cap can only decrease");
        require(cap >= seasonMinted[season], "cap below minted");
        seasonCap[season] = cap;
        emit SeasonCapSet(season, cap);
    }

    /// @notice Mint settled rewards for a season. Called by the settlement
    ///         service in batches after the internal ledger has vested them.
    function settleMint(uint256 season, address to, uint256 amount) external onlyRole(SETTLER_ROLE) {
        uint256 remainingSeason = seasonCap[season] - seasonMinted[season];
        if (amount > remainingSeason) revert SeasonCapExceeded(season, amount, remainingSeason);
        uint256 remainingTotal = maxSupply - totalSupply();
        if (amount > remainingTotal) revert MaxSupplyExceeded(amount, remainingTotal);

        seasonMinted[season] += amount;
        _mint(to, amount);
        emit SeasonMint(season, to, amount);
    }

    function pause() external onlyRole(PAUSER_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(PAUSER_ROLE) {
        _unpause();
    }

    function _update(address from, address to, uint256 value)
        internal
        override(ERC20, ERC20Pausable)
    {
        super._update(from, to, value);
    }
}
