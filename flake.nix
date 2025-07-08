{
  description = "A Nix-flake-based Go dev env";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    aliases.url = "github:teodord25/dotfiles?dir=flakes/git-aliases";
  };

  outputs =
    {
      self,
      nixpkgs,
      aliases,
      ...
    }:
    let
      goVersion = "1_24"; # bump when you need
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      overlays.default = final: prev: {
        go = prev."go_${goVersion}";
      };

      devShells = forAllSystems (
        system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [
              self.overlays.default
              aliases.overlays.default
            ];
          };
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gotools
              golangci-lint
              gopls # language-server
            ];

            shellHook = ''
              # shared aliases from your alias flake
              ${pkgs.sharedAliases}

              # Go module / GOPATH hygiene
              export GOPATH=$PWD/.go
              export GO111MODULE=on

              echo "Go ${goVersion} dev-shell ready 🚀"
            '';
          };
        }
      );
    };
}
