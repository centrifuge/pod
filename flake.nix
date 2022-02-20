{
  description = "go-centrifuge";

  inputs.nixpkgs.url = github:NixOS/nixpkgs/nixos-21.11;

  outputs = { self, nixpkgs }:
    let
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
      name = "go-centrifuge";
      version = "2.0";
      system = "x86_64-linux";
    in
    {

      packages.${system} = {
        go-centrifuge = pkgs.buildGoModule rec {
          pname = name;
          inherit version;

          src = ./.;

          vendorSha256 = "sha256-Yr1lxkeW9rvOR+tLn9YBE682hgatsLGRqsalhcz+r9Y=";
        };

        dockerContainer =
          pkgs.dockerTools.buildImage {
            name = "centrifugeio/${name}";
            tag = version;

            contents = self.defaultPackage.${system};

            config = {
              Env = [
                "PATH=/bin:$PATH"
              ];
              ExposedPorts = {
                "30333/tcp" = { };
                "9933/tcp" = { };
                "9944/tcp" = { };
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
