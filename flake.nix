{
  description = "go-centrifuge";

  inputs = {
    nixpkgs.url = github:NixOS/nixpkgs/nixos-21.11;
    gitignore = {
      url = github:hercules-ci/gitignore.nix;
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, gitignore }:
    let
      name = "go-centrifuge";
      version = "2.0";

      system = "x86_64-linux";

      pkgs = nixpkgs.legacyPackages.${system};

      srcFilter = src:
        let
          srcIgnored = gitignore.lib.gitignoreFilter src;
          ignoreList = [
            ".envrc"
            ".github"
            "CODE_OF_CONDUCT.md"
            "README.md"
            "codecov.yml"
            "flake.lock"
            "flake.nix"
          ];
        in
        path: type:
          srcIgnored path type
          && builtins.all (name: builtins.baseNameOf path != name) ignoreList;
    in
    {
      packages.${system} = {
        go-centrifuge = pkgs.buildGoModule {
          pname = name;
          inherit version;

          src = pkgs.lib.cleanSourceWith {
            src = ./.;
            filter = srcFilter ./.;
            name = "${name}-source";
          };

          vendorSha256 = "sha256-Yr1lxkeW9rvOR+tLn9YBE682hgatsLGRqsalhcz+r9Y=";
        };

        dockerImage =
          let
            # This evaluates to the first 6 digits of the git hash of this repo's HEAD
            # commit, or to "dirty" if there are uncommitted changes.
            commit-substr = builtins.substring 0 6 (self.rev or "dirty");

            tag = "${version}-${commit-substr}";
          in
          pkgs.dockerTools.buildLayeredImage {
            name = "centrifugeio/${name}";
            inherit tag;
            created = builtins.substring 0 8 self.lastModifiedDate;

            contents = [
              pkgs.busybox
              self.defaultPackage.${system}
            ];

            config = {
              ExposedPorts = {
                # api
                "8082/tcp" = { };
                # p2p
                "38202/tcp" = { };
              };
              Volumes = {
                "/data" = { };
              };
              Entrypoint = [ "centrifuge" ];
            };
          };
      };

      defaultPackage.${system} = self.packages.${system}.go-centrifuge;

      apps.${system}.centrifuge = {
        type = "app";
        program = "${self.defaultPackage.${system}}/bin/centrifuge";
      };

      defaultApp.${system} = self.apps.${system}.centrifuge;

      devShell.${system} = pkgs.mkShellNoCC {
        buildInputs = [ self.defaultPackage.${system} ];
      };
    };
}
