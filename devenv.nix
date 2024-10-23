{ pkgs, lib, config, inputs, ... }:

{
  packages = [
    pkgs.gnumake
  ];
  
  # https://devenv.sh/languages/
  languages.go.enable = true;

  # See full reference at https://devenv.sh/reference/options/
}
