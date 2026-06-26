#!/bin/sh
set -e

KREW_ROOT="${KREW_ROOT:-/root/.krew}"
export PATH="${KREW_ROOT}/bin:${PATH}"

# Shell identity (prompt + welcome banner)
WS_USER="${WORKSTATION_USER:-stormtrooper}"
WS_HOST="${WORKSTATION_HOST:-deathstar}"
export WORKSTATION_USER="$WS_USER"
export WORKSTATION_HOST="$WS_HOST"
hostname "$WS_HOST" 2>/dev/null || true
mkdir -p /etc/krew-workstation
cat > /etc/krew-workstation/shell.env <<EOF
export WORKSTATION_USER='$WS_USER'
export WORKSTATION_HOST='$WS_HOST'
export PS1='${WS_USER}@${WS_HOST}:\w\$ '
EOF
if ! grep -q 'krew-workstation shell.env' /root/.bashrc 2>/dev/null; then
  touch /root/.bashrc
  echo '[ -f /etc/krew-workstation/shell.env ] && . /etc/krew-workstation/shell.env' >> /root/.bashrc
fi

# If krew binary missing (e.g. fresh volume overwrote image), install krew
if ! command -v krew >/dev/null 2>&1; then
  echo "[entrypoint] Installing krew..."
  OS="linux"
  ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')"
  KREW_TAR="krew-${OS}_${ARCH}.tar.gz"
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW_TAR}"
  tar zxvf "${KREW_TAR}"
  ./"krew-${OS}_${ARCH}" install krew
  rm -f "${KREW_TAR}" "krew-${OS}_${ARCH}"
fi

# SSH key for node access (generate if not exists)
mkdir -p /root/.ssh
if [ ! -f /root/.ssh/id_rsa ]; then
  ssh-keygen -t rsa -N "" -f /root/.ssh/id_rsa -q
fi

# Bash completion for kubectl and krew (includes plugin completion)
if ! grep -q 'kubectl-completion' /root/.bashrc 2>/dev/null; then
  touch /root/.bashrc
  kubectl completion bash > /root/.kubectl-completion.bash 2>/dev/null || true
  kubectl krew completion bash > /root/.krew-completion.bash 2>/dev/null || true
  {
    echo ''
    echo '# kubectl and krew completion'
    echo '[ -f /root/.kubectl-completion.bash ] && source /root/.kubectl-completion.bash'
    echo '[ -f /root/.krew-completion.bash ] && source /root/.krew-completion.bash'
  } >> /root/.bashrc
fi

# krew-workstation shell aliases (avoid injecting via PTY — prevents echo in terminal)
if ! grep -q 'krew-workstation aliases' /root/.bashrc 2>/dev/null; then
  {
    echo ''
    echo '# krew-workstation aliases'
    echo "alias k=kubectl"
    echo "alias kk='kubectl krew'"
  } >> /root/.bashrc
fi

# Update plugin index
kubectl krew update 2>/dev/null || true

# Local dev plugins bind-mounted at /opt/local-plugins/<name> (e.g. rk2s)
. /install-local-plugins.sh
install_local_krew_plugins

# Local CLIs from .local-plugins/<name>/bin/ (rk9s, rancher-polymorph, ...)
. /install-local-binaries.sh
install_local_binaries

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
