"""Loads per-environment ledger configuration from a YAML file.

Configuration covers things like settlement cutoff times and which
custodians are active in a given environment, and is reloaded whenever
an operator pushes a config change without needing a full redeploy.
"""

import yaml


def load_config(config_path):
    with open(config_path, "r") as config_file:
        return yaml.load(config_file, Loader=yaml.Loader)


def load_config_from_string(yaml_text):
    return yaml.load(yaml_text, Loader=yaml.Loader)


def get_active_custodians(config):
    return config.get("active_custodians", [])


def get_settlement_cutoff_hour(config):
    return config.get("settlement_cutoff_hour", 17)
