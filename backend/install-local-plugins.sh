#!/bin/sh
# Install krew plugins bind-mounted from /opt/local-plugins/<name> (local dev).

install_local_krew_plugins() {
  base="/opt/local-plugins"

  if [ ! -d "$base" ]; then
    return 0
  fi

  for dir in "$base"/*; do
    [ -d "$dir" ] || continue

    plugin="$(basename "$dir")"
    manifest=""

    for f in \
      "$dir/.krew.yaml" \
      "$dir/plugin.yaml" \
      "$dir/manifest.yaml" \
      "$dir/manifests/${plugin}.yaml" \
      "$dir/manifests/plugin.yaml"
    do
      if [ -f "$f" ]; then
        manifest="$f"
        break
      fi
    done

    if [ -z "$manifest" ]; then
      if [ "$plugin" = "rk9s" ] || [ "$plugin" = "rancher-polymorph" ]; then
        continue
      fi
      if [ -d "$dir/bin" ]; then
        continue
      fi
      echo "[entrypoint] skip local plugin ${plugin}: no krew manifest found"
      continue
    fi

    if kubectl krew list 2>/dev/null | grep -qx "$plugin"; then
      echo "[entrypoint] local plugin ${plugin} already installed"
      continue
    fi

    echo "[entrypoint] installing local krew plugin ${plugin} from ${manifest}"

    if [ -f "$dir/Makefile" ] && command -v make >/dev/null 2>&1; then
      (cd "$dir" && make build) || echo "[entrypoint] make build failed for ${plugin}"
    fi

    if kubectl krew install --manifest="$manifest"; then
      echo "[entrypoint] installed local plugin ${plugin}"
    else
      echo "[entrypoint] failed to install local plugin ${plugin}"
    fi
  done
}
