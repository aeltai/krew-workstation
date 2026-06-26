#!/bin/sh
# Install prebuilt linux binaries from /opt/local-plugins/<name>/bin/<name>

install_local_binary() {
  name="$1"
  dest="/usr/local/bin/$name"

  for candidate in \
    "/opt/local-plugins/${name}/bin/${name}" \
    "/opt/local-plugins/${name}/${name}-linux-amd64" \
    "/opt/local-plugins/${name}/${name}"
  do
    if [ ! -f "$candidate" ] || [ ! -s "$candidate" ]; then
      continue
    fi

    magic="$(head -c 4 "$candidate" | od -An -tx1 | tr -d ' \n')"
    if [ "$magic" != "7f454c46" ]; then
      continue
    fi

    cp "$candidate" "$dest"
    chmod +x "$dest"
    echo "[entrypoint] installed ${name} from ${candidate}"
    return 0
  done

  if [ -x "$dest" ]; then
    echo "[entrypoint] using ${name} already in image"
    return 0
  fi

  echo "[entrypoint] ${name} not installed — run: ./scripts/build-local-cli-for-docker.sh ${name}"
}

install_local_binaries() {
  install_local_binary rk9s
  install_local_binary rancher-polymorph
}
