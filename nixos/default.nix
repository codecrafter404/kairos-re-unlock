# Compatibility wrapper for non-flake users.
# Usage in configuration.nix:
#   imports = [ (import /path/to/this/repo/nixos) ];
#
# Note: the droplet and CLI packages must be provided via an overlay or
# the `services.remote-unlock.package` option.
(import ./module.nix null)
