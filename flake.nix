{
  description = "Go API Development Environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config = {
            allowUnfree = true;
          };
        };
      in {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_24
            gopls
            gotools        # includes goimports
            golangci-lint  # linting
            delve          # debugger
            air            # hot reload for development

            # Database tools
            postgresql_16
            goose          # database migrations

            # Container tools
            docker
            docker-compose

            # Utilities
            jq
            gnumake
            just
          ];

          shellHook = ''
            echo "Go API Development Environment"
            echo "Available tools:"
            echo "- Go $(go version)"
            echo "- PostgreSQL $(postgres --version | head -1)"
            echo "- Docker $(docker --version)"
            echo "- Just $(just --version)"
            echo ""
            echo "To see available commands, run: just"
            echo ""

            export GOBIN="$PWD/bin"
            export PATH="$GOBIN:$PATH"

            # Install swag if not present
            if ! command -v swag &> /dev/null; then
              echo "Installing swag..."
              go install github.com/swaggo/swag/cmd/swag@latest
            fi
          '';
        };
      });
}
