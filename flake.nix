{
  description = "NixOS Remote Unlock – remotely unlock LUKS-encrypted partitions over HTTP or P2P";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          re-unlock-droplet = pkgs.buildGoModule {
            pname = "re-unlock-droplet";
            version = self.shortRev or self.dirtyShortRev or "dev";
            src = ./.;
            subPackages = [ "droplet" ];
            # After the first build, replace this placeholder with the real hash
            # printed by the failed build. For example:
            #   nix build .#re-unlock-droplet 2>&1 | grep 'got:'
            vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
            CGO_ENABLED = 0;
            postInstall = ''
              mv $out/bin/droplet $out/bin/re-unlock-droplet
            '';
            meta = {
              description = "Remote LUKS unlock daemon (droplet)";
              mainProgram = "re-unlock-droplet";
            };
          };

          re-unlock-cli = pkgs.buildGoModule {
            pname = "re-unlock-cli";
            version = self.shortRev or self.dirtyShortRev or "dev";
            src = ./.;
            subPackages = [ "client" ];
            # Same as above – update after first build.
            vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
            CGO_ENABLED = 0;
            postInstall = ''
              mv $out/bin/client $out/bin/re-unlock-cli
            '';
            meta = {
              description = "Remote LUKS unlock CLI client";
              mainProgram = "re-unlock-cli";
            };
          };

          default = self.packages.${system}.re-unlock-droplet;
        });

      nixosModules.default = import ./nixos/module.nix self;

      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [ go gopls gotools ];
          };
        });
    };
}
