#!/bin/sh
set -e

# Use /opt/krew for non-root compatibility (Rancher enforces runAsNonRoot)
KREW_ROOT="${KREW_ROOT:-/opt/krew}"
export KREW_ROOT
export HOME="${HOME:-/opt/krew-workstation}"
export PATH="${KREW_ROOT}/bin:${PATH}"
CONFIG_DIR="${CONFIG_DIR:-/opt/krew-workstation}"
mkdir -p "${CONFIG_DIR}"
mkdir -p "${KREW_ROOT}"

# If krew not present (e.g. fresh PVC at /opt/krew), copy from image. Skip if volume already has krew (e.g. root-owned from prior run).
if [ ! -f "${KREW_ROOT}/bin/kubectl-krew" ] && [ -d /usr/local/share/krew-default ]; then
  echo "[entrypoint] Restoring krew from image into ${KREW_ROOT}..."
  cp -a /usr/local/share/krew-default/. "${KREW_ROOT}/" || true
elif ! command -v krew >/dev/null 2>&1 && [ ! -d /usr/local/share/krew-default ]; then
  echo "[entrypoint] ERROR: /usr/local/share/krew-default missing; rebuild image."
  exit 1
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

# Install ssh-jump first (sync) — user expects it immediately. rk9s is preinstalled as standalone binary.
kubectl krew list 2>/dev/null | grep -q "^ssh-jump$" || kubectl krew install ssh-jump 2>/dev/null || true

# Install other default plugins in background (browse-pvc needs gcompat for glibc binary)
(
  for p in stern lineage get-all crust-gather browse-pvc; do
    kubectl krew list 2>/dev/null | grep -q "^${p}$" || kubectl krew install "$p" 2>/dev/null || true
  done
) &

exec krew-manager
