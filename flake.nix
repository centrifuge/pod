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

      gitignoreFilter = gitignore.lib.gitignoreFilter;

      srcFilter = src:
        let
          srcIgnored = gitignoreFilter src;
        in
        path: type:
          let
            p = builtins.baseNameOf path;
          in
          srcIgnored path type
          || (p == "flake.nix");
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

        dockerImage = pkgs.dockerTools.buildLayeredImage {
          name = "centrifugeio/${name}";
          tag = version;

          contents = [ pkgs.busybox self.defaultPackage.${system} ];

          config = {
            # TODO what ports do we actually need?
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
    };
}
