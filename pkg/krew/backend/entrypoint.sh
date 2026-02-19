#!/bin/sh
set -e

# Use /opt/krew for non-root compatibility (Rancher enforces runAsNonRoot)
KREW_ROOT="${KREW_ROOT:-/opt/krew}"
export KREW_ROOT
export PATH="${KREW_ROOT}/bin:${PATH}"
CONFIG_DIR="${CONFIG_DIR:-/opt/krew-workstation}"
mkdir -p "${CONFIG_DIR}"
mkdir -p "${KREW_ROOT}"

# If krew binary missing (e.g. fresh PVC mounted at /opt/krew), copy from image — no runtime installer (avoids /root/.krew permission denied under runAsNonRoot)
if ! command -v krew >/dev/null 2>&1; then
  echo "[entrypoint] Restoring krew from image into ${KREW_ROOT}..."
  if [ -d /usr/local/share/krew-default ]; then
    cp -a /usr/local/share/krew-default/. "${KREW_ROOT}/"
  else
    echo "[entrypoint] ERROR: /usr/local/share/krew-default missing; rebuild image."
    exit 1
  fi
fi

# SSH key for node access (generate if not exists)
mkdir -p "${CONFIG_DIR}/.ssh"
if [ ! -f "${CONFIG_DIR}/.ssh/id_rsa" ]; then
  ssh-keygen -t rsa -N "" -f "${CONFIG_DIR}/.ssh/id_rsa" -q
fi

# Bash completion for kubectl and krew
BASHRC="${CONFIG_DIR}/.bashrc"
if ! grep -q 'kubectl-completion' "${BASHRC}" 2>/dev/null; then
  touch "${BASHRC}"
  kubectl completion bash > "${CONFIG_DIR}/.kubectl-completion.bash" 2>/dev/null || true
  kubectl krew completion bash > "${CONFIG_DIR}/.krew-completion.bash" 2>/dev/null || true
  {
    echo ''
    echo '# kubectl and krew completion'
    echo "[ -f ${CONFIG_DIR}/.kubectl-completion.bash ] && source ${CONFIG_DIR}/.kubectl-completion.bash"
    echo "[ -f ${CONFIG_DIR}/.krew-completion.bash ] && source ${CONFIG_DIR}/.krew-completion.bash"
  } >> "${BASHRC}"
fi

# Update plugin index
kubectl krew update 2>/dev/null || true

# Install k9s and ssh-jump first (sync) — user expects them immediately
kubectl krew list 2>/dev/null | grep -q "^k9s$" || kubectl krew install k9s 2>/dev/null || true
kubectl krew list 2>/dev/null | grep -q "^ssh-jump$" || kubectl krew install ssh-jump 2>/dev/null || true

# Install other default plugins in background
(
  for p in stern lineage get-all crust-gather; do
    kubectl krew list 2>/dev/null | grep -q "^${p}$" || kubectl krew install "$p" 2>/dev/null || true
  done
) &

exec krew-manager
