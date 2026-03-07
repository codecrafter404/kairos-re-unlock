# NixOS Remote Unlock

Remotely unlock LUKS-encrypted partitions on a NixOS system. The service provides network connectivity during early boot (initrd) and receives the decryption password through multiple channels:

- **HTTP** – POST to the droplet on port `:505`
- **P2P** – EdgeVPN / go-nodepair based peer-to-peer transport
- **HTTP Pull** – the droplet polls a remote server for the payload

An optional WiFi configuration lets you unlock machines that are only reachable over wireless.

The project also ships WireGuard support and Discord webhook notifications.

## Quick start (NixOS flake)

Add this repository as a flake input and import the module:

```nix
# flake.nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
    remote-unlock.url = "github:codecrafter404/kairos-re-unlock";
  };

  outputs = { nixpkgs, remote-unlock, ... }: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        remote-unlock.nixosModules.default
        ./configuration.nix
      ];
    };
  };
}
```

Then enable the service in your `configuration.nix`:

```nix
services.remote-unlock = {
  enable = true;
  edgevpnToken = "b3RwOg...==";
  publicKey = builtins.readFile ./keys/client_pub.pem;
  privateKey = builtins.readFile ./keys/droplet_priv.pem;
  # Optional settings
  # discordWebhook = "https://discord.com/api/webhooks/...";
  # ntpServer = "time.cloudflare.com";
  # httpPull = [ "10.0.0.5:505" ];
  # luksDevice = "sda2";
};

# Configure LUKS as usual
boot.initrd.luks.devices."cryptroot" = {
  device = "/dev/disk/by-uuid/...";
};
```

### Generate keys and tokens

```bash
nix run github:codecrafter404/kairos-re-unlock#re-unlock-cli -- new
```

This outputs both the droplet configuration block and the client keys.

## WiFi in initrd

To bring up WiFi before the LUKS unlock, enable the WiFi option and supply the required kernel modules and firmware for your hardware:

```nix
services.remote-unlock.wifi = {
  enable = true;
  interfaces = [ "wlan0" ];
  kernelModules = [ "brcmfmac" "brcmutil" ];
};
```

You will also need to configure `wpa_supplicant` in the initrd.  See the NixOS manual section on `boot.initrd.network` for details.

## Configuration file

The droplet reads its YAML configuration from the first file it finds in the following directories (in order):

1. `/etc/reunlock/`
2. `/oem/`
3. `/sysroot/oem/`
4. `/tmp/oem/`

When the NixOS module is enabled the config is generated automatically at `/etc/reunlock/config.yaml`.

The expected YAML format:

```yaml
kcrypt:
  remote_unlock:
    edgevpn_token: b3RwOg...==
    public_key: |
      -----BEGIN PUBLIC KEY-----
      ...
      -----END PUBLIC KEY-----
    private_key: |
      -----BEGIN PRIVATE KEY-----
      ...
      -----END PRIVATE KEY-----
    # Optional
    discord_webhook: https://discord.com/api/webhooks/...
    http_pull:
      - 10.0.0.5:505
    ntp_server: time.cloudflare.com
    debug:
      enabled: false
      log_level: -1
      password: ""
      bypass_password_test: false
```

## CLI usage

| Command | Description |
|---|---|
| `re-unlock-cli new` | Generate keys, token & config |
| `re-unlock-cli token` | Generate an EdgeVPN token |
| `re-unlock-cli unlock` | Send password via P2P |
| `re-unlock-cli unlock-http` | Send password via HTTP POST |
| `re-unlock-cli unlock-serve` | Serve the password via HTTP GET |
| `re-unlock-cli logs -i <ip>` | Fetch logs from the droplet (debug mode) |
| `re-unlock-cli version` | Print version |

## Droplet modes

The droplet binary supports two modes:

- `re-unlock-droplet discovery.password` – Kairos-compatible mode. Reads partition JSON from stdin, outputs a JSON `EventResponse` to stdout.
- `re-unlock-droplet unlock [device]` – NixOS mode. Optionally validates the password against the given LUKS device and prints the raw password to stdout.

Running the droplet without arguments prints the current configuration for debugging.

## Notifications

Add a Discord webhook for status notifications:

```nix
services.remote-unlock.discordWebhook = "https://discord.com/api/webhooks/...";
```

## Debug mode

> **WARNING:** Debug options leak the private and public key on the `/logs` endpoint. Do not enable in production.

```nix
services.remote-unlock.debug = {
  enable = true;
  logLevel = -1;
  password = "supersecurepassword";
  bypassPasswordTest = false;
};
```

## Naming

- **droplet** – the daemon running on the NixOS machine, waiting for the unlock password
- **client** – the CLI tool used to send the password

## Building

### With Nix

```bash
nix build .#re-unlock-droplet
nix build .#re-unlock-cli
```

### With Go

```bash
CGO_ENABLED=0 go build -o re-unlock-droplet ./droplet/main.go
CGO_ENABLED=0 go build -o re-unlock-cli ./client/main.go
```

### CI / Automation

The GitHub Actions workflow builds the CLI and droplet binaries for multiple platforms and creates a release. A `nix flake check` step verifies the Nix flake.

