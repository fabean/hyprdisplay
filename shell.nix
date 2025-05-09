{ pkgs ? import <nixpkgs> {} }:

(pkgs.buildFHSUserEnv {
  name = "hyprdisplay";
  targetPkgs = pkgs: with pkgs; [
    go
  ];
  profile = ''
    export GO111MODULE=on
  '';
}).env