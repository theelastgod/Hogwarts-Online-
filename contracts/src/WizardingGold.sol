// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC20Burnable} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Burnable.sol";
import {ERC20Permit} from "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";

/// @title Wizarding Gold (WGLD)
/// @notice Fixed-supply token for Hogwarts Online, launched via launchpad.
/// @dev The entire supply is minted once to a distributor at deployment and
///      split into vesting contracts and the RewardsVault by the deployer
///      script. There is no mint function, no owner, and no pause: launchpad
///      and exchange due diligence expect a token that cannot be changed
///      after TGE. Supply can only decrease via burn.
contract WizardingGold is ERC20, ERC20Burnable, ERC20Permit {
    /// @notice 1,000,000,000 WGLD.
    uint256 public constant TOTAL_SUPPLY = 1_000_000_000e18;

    constructor(address distributor) ERC20("Wizarding Gold", "WGLD") ERC20Permit("Wizarding Gold") {
        require(distributor != address(0), "distributor zero");
        _mint(distributor, TOTAL_SUPPLY);
    }
}
