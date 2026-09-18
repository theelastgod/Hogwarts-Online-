// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";

/// @title Rewards Vault
/// @notice Holds the player-rewards allocation of WGLD and releases it to the
///         settlement wallet under per-season caps.
/// @dev Mirrors the emission cap in services/ledger. The settlement worker
///      draws from this vault to fund player withdrawals. Caps can only be
///      lowered once set, and the admin is expected to be a timelocked
///      multisig. Seasons are 90 days from `genesis`.
contract RewardsVault is AccessControl, Pausable {
    using SafeERC20 for IERC20;

    bytes32 public constant SETTLER_ROLE = keccak256("SETTLER_ROLE");
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");

    IERC20 public immutable token;
    uint64 public immutable genesis;
    uint64 public constant SEASON_LENGTH = 90 days;

    /// @notice Initial season cap; decays 4% per season unless overridden lower.
    uint256 public constant INITIAL_SEASON_CAP = 20_000_000e18;
    uint256 public constant DECAY_BPS = 400; // 4% per season

    mapping(uint256 => uint256) public capOverride; // 0 = use schedule
    mapping(uint256 => uint256) public released;

    event SeasonCapLowered(uint256 indexed season, uint256 cap);
    event Released(uint256 indexed season, address indexed to, uint256 amount);

    error SeasonCapExceeded(uint256 season, uint256 requested, uint256 remaining);
    error CapCanOnlyDecrease(uint256 season, uint256 current, uint256 requested);

    constructor(IERC20 token_, address admin, uint64 genesis_) {
        token = token_;
        genesis = genesis_;
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(PAUSER_ROLE, admin);
    }

    /// @notice Current season index, 0-based, from genesis.
    function currentSeason() public view returns (uint256) {
        if (block.timestamp < genesis) return 0;
        return (block.timestamp - genesis) / SEASON_LENGTH;
    }

    /// @notice Scheduled cap for a season: INITIAL * 0.96^season, or the override.
    function seasonCap(uint256 season) public view returns (uint256) {
        if (capOverride[season] != 0) return capOverride[season];
        uint256 cap = INITIAL_SEASON_CAP;
        for (uint256 i = 0; i < season; i++) {
            cap = cap * (10_000 - DECAY_BPS) / 10_000;
        }
        return cap;
    }

    /// @notice Lower a season's cap. Cannot raise above the schedule or a prior override.
    function lowerSeasonCap(uint256 season, uint256 cap) external onlyRole(DEFAULT_ADMIN_ROLE) {
        uint256 current = seasonCap(season);
        if (cap >= current) revert CapCanOnlyDecrease(season, current, cap);
        require(cap >= released[season], "cap below released");
        capOverride[season] = cap;
        emit SeasonCapLowered(season, cap);
    }

    /// @notice Release rewards for the current season to the settlement wallet.
    function release(address to, uint256 amount) external onlyRole(SETTLER_ROLE) whenNotPaused {
        // Genesis is set to game launch, so a launchpad TGE before that date
        // cannot release rewards until there is a game to earn them in.
        require(block.timestamp >= genesis, "before genesis");
        uint256 season = currentSeason();
        uint256 remaining = seasonCap(season) - released[season];
        if (amount > remaining) revert SeasonCapExceeded(season, amount, remaining);
        released[season] += amount;
        token.safeTransfer(to, amount);
        emit Released(season, to, amount);
    }

    function pause() external onlyRole(PAUSER_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(PAUSER_ROLE) {
        _unpause();
    }
}
