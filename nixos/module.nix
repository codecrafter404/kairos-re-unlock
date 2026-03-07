flake:
{ config, lib, pkgs, ... }:
let
  cfg = config.services.remote-unlock;
  yamlFormat = pkgs.formats.yaml { };

  # Build the configuration YAML expected by the droplet
  configContent = {
    kcrypt.remote_unlock = {
      edgevpn_token = cfg.edgevpnToken;
      public_key = cfg.publicKey;
      private_key = cfg.privateKey;
    } // lib.optionalAttrs (cfg.discordWebhook != null) {
      discord_webhook = cfg.discordWebhook;
    } // lib.optionalAttrs (cfg.httpPull != [ ]) {
      http_pull = cfg.httpPull;
    } // lib.optionalAttrs (cfg.ntpServer != null) {
      ntp_server = cfg.ntpServer;
    } // lib.optionalAttrs cfg.debug.enable {
      debug = {
        enabled = true;
        log_level = cfg.debug.logLevel;
      } // lib.optionalAttrs (cfg.debug.password != null) {
        password = cfg.debug.password;
      } // lib.optionalAttrs cfg.debug.bypassPasswordTest {
        bypass_password_test = true;
      };
    };
  };

  configFile = yamlFormat.generate "re-unlock-config.yaml" configContent;
  dropletPkg = cfg.package;
in
{
  options.services.remote-unlock = {
    enable = lib.mkEnableOption "Remote LUKS unlock service";

    package = lib.mkPackageOption pkgs "re-unlock-droplet" {
      default = flake.packages.${pkgs.stdenv.hostPlatform.system}.re-unlock-droplet;
    };

    edgevpnToken = lib.mkOption {
      type = lib.types.str;
      description = "EdgeVPN token for P2P communication.";
    };

    publicKey = lib.mkOption {
      type = lib.types.str;
      description = "PEM-encoded RSA public key of the client.";
    };

    privateKey = lib.mkOption {
      type = lib.types.str;
      description = "PEM-encoded RSA private key of the droplet.";
    };

    discordWebhook = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      description = "Optional Discord webhook URL for notifications.";
    };

    httpPull = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [ ];
      description = "List of HTTP addresses to poll for the unlock payload.";
    };

    ntpServer = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      description = "NTP server for time synchronisation (default: time.cloudflare.com).";
    };

    luksDevice = lib.mkOption {
      type = lib.types.nullOr lib.types.str;
      default = null;
      example = "sda2";
      description = "Device name for LUKS password validation (without /dev/ prefix). If null, validation is skipped.";
    };

    wifi = {
      enable = lib.mkEnableOption "WiFi connectivity in initrd for remote unlock";

      interfaces = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ "wlan0" ];
        description = "Wireless interfaces to bring up in the initrd.";
      };

      kernelModules = lib.mkOption {
        type = lib.types.listOf lib.types.str;
        default = [ ];
        example = [ "brcmfmac" "brcmutil" ];
        description = "Extra kernel modules needed for WiFi in initrd.";
      };

      firmware = lib.mkOption {
        type = lib.types.listOf lib.types.path;
        default = [ ];
        description = "Firmware files to include in the initrd for WiFi.";
      };
    };

    wireguard.enable = lib.mkEnableOption "WireGuard support";

    debug = {
      enable = lib.mkEnableOption "Debug mode (WARNING: leaks keys on /logs endpoint)";

      logLevel = lib.mkOption {
        type = lib.types.int;
        default = 0;
        description = "Zerolog log level integer.";
      };

      password = lib.mkOption {
        type = lib.types.nullOr lib.types.str;
        default = null;
        description = "Debug password – bypasses encryption if set.";
      };

      bypassPasswordTest = lib.mkOption {
        type = lib.types.bool;
        default = false;
        description = "Skip LUKS password validation (may lead to broken systems).";
      };
    };
  };

  config = lib.mkIf cfg.enable {
    # Place config on disk for the droplet to find
    environment.etc."reunlock/config.yaml".source = configFile;

    # Networking in initrd
    boot.initrd.network.enable = true;

    # WiFi support in initrd
    boot.initrd.availableKernelModules = lib.mkIf cfg.wifi.enable cfg.wifi.kernelModules;

    boot.initrd.network.postCommands = lib.mkAfter ''
      # Run remote unlock droplet and write password for LUKS
      mkdir -p /crypt-ramfs
      ${dropletPkg}/bin/re-unlock-droplet unlock ${lib.optionalString (cfg.luksDevice != null) cfg.luksDevice} > /crypt-ramfs/passphrase 2>/tmp/re-unlock-initrd.log &
    '';

    # WireGuard
    networking.wireguard.enable = lib.mkIf cfg.wireguard.enable true;
    boot.extraModulePackages = lib.mkIf cfg.wireguard.enable [
      config.boot.kernelPackages.wireguard
    ];

    # IP forwarding for WireGuard
    boot.kernel.sysctl = lib.mkIf cfg.wireguard.enable {
      "net.ipv4.ip_forward" = 1;
      "net.ipv6.conf.all.forwarding" = 1;
      "net.ipv6.conf.default.forwarding" = 1;
    };

    # Useful utilities
    environment.systemPackages = [
      dropletPkg
      pkgs.cryptsetup
    ] ++ lib.optionals cfg.wireguard.enable [
      pkgs.wireguard-tools
    ];
  };
}
