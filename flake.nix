{
  description = "Development environment with Telepresence";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... } @ inputs:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config = {
            allowUnfree = false; # Depending on your license preferences
          };
        };
        telepresence = pkgs.telepresence2;
      in {
        devShell = pkgs.mkShell {
          buildInputs = [
            telepresence
            # ... any other dependencies
          ];

          # Set any environment variables or shell hooks that you need for your development
          # environment here
          # shellHook = ''
          #   export SOME_ENV_VAR="value"
          # '';
        };
      }
    );
}
