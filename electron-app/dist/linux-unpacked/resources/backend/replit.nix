{pkgs}: {
  deps = [
    pkgs.docker
    pkgs.postgresql
    pkgs.pkg-config
    pkgs.gcc
    pkgs.sqlite
  ];
}
