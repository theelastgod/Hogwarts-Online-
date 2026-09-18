"""Hogwarts Online economy simulation.

Models WGLD (Wizarding Gold) emissions against sinks over seasons to check
that per-token value stays within the bounds set in DEVELOPMENT_PLAN.md.
Pure standard library so it runs anywhere.
"""

from .model import EconomyConfig, SeasonResult, simulate

__all__ = ["EconomyConfig", "SeasonResult", "simulate"]
