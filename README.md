# Kairos-Remote-Unlock
This custom alpine flavour enables remote unlocking of the luks encrypted partitions. It allows connectivity over wifi. It mainly receives the password through two channels:
- Http: an http request to the server on port `:505`
- Using kairos nodepair based on P2P -> sometimes really unreliable

For my private usage the wifi functionallity only works on some firmware, which is automatically included into the modules in `kernel/net/wireless` and `kernel/drivers/net/wireless`. Furthermore it includes the `brcmfmac43455-sdio.raspberrypi,4-model-b` firmware for the raspberry pi 4b.


The wifi will automatically turned off after initramfs in order for the main system to allow for wifi connectivity there.


The image also includes wireguard.
## How to add a wifi connectivity
For initial wifi connectivity during initramfs (during decryption), just put a `wpa.conf` in `/oem/wpa.conf`. It can be generated using:

```bash
wpa_passphrase SSID SuperSecurePassword > /oem/wpa.conf
```

## Usage
### Setup
If you enable encryption you have to set up the following parts in the config file (in OEM, which is unencrypted at rest):
```yaml
kcrypt:
   remote_unlock:
      edgevpn_token: b3RwOgo<snip>==
      # Public Key of the client
      public_key: |
         -----BEGIN PUBLIC KEY-----
         MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnQXyiHLnHgh7ctM6kmG4
         <snip>
         KepPymg6mdt8dn405JGI+lqmBiuq59Zp5W5sI7akeP9joMyi6+8OFvc8Zstrh7go
         ZQIDAQAB
         -----END PUBLIC KEY-----
      # Private Key of Droplet
      private_key: |
         -----BEGIN PRIVATE KEY-----
         MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDpKvh1oEA644EP
         <snip>
         gGpi0iY7JnClU1J0pJ6Uts4=
         -----END PRIVATE KEY-----
```
This configuration can be generated using
```bash
kairos-re-unlock new
```
This command also outputs the corresponding public and private keys to be used for decryption.
### Debug mode
- to test the password receiving functionallity run `echo {} | /system/discovery/kcrypt-discovery-re-unlock discovery.password`
- just running `/system/discovery/kcrypt-discovery-re-unlock` gives the config as output

### Usage of cli
- for the usage of the cli, you can simply run the cli

### Notification
In order to allow discord notifications add the `discord_webhook` parameter:
```yaml
kcrypt:
   remote_unlock:
      discord_webhook: https://discord.com/api/webhooks/<snap>
```
### Debug
The're some debug options. Debug mode has to be enabled before the other options may be used.

> WARNING: Debug options should not be enabled in production as they leak the private and public key of the droplet on the `/logs` endpoint.

```yaml
kcrypt:
   remote_unlock:
      debug:
         enabled: false
         # Integer log level
         log_level: -1
         # Provides the password and therefore there is no need to enable encryption
         password: supersecurepassword
         # Bypasses the password validation which leads to broken systems, if a wrong password is provided
         bypass_password_test: false
```

## Naming
- The decryption is handled by the `droplet` on the kairos-machine
- the `client` sends the password

## Building / Automation
The build & upgrade process of the image is automated. (using github actions)
### Building
- the build process is automated for arm (raspberry pi) and x86 maschines
- It utilizes the kairos factory action
### Upgrade
- The upgrade process pulls the current versions from the main karios repo and applies them to the github action
- also the factory action is automatically updated
- the current kubernetes version is directly from k3s github repo
- **WARNING:** only this action will merge the Dockerfile_ext and base file for the new docker file. Before running build, after modifiying the dockerfile, you should run this action
