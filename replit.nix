{pkgs}: {
  deps = [
    pkgs.postgresql
    pkgs.pkg-config
    pkgs.gcc
    pkgs.sqlite
  ];
}
