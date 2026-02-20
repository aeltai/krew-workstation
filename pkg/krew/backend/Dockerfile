FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o krew-manager .

# Build rk9s (SUSE/Rancher k9s fork - rancher, fleet, rke2-k3s, longhorn, harvester, kubewarden)
FROM golang:1.25-alpine AS rk9s-builder
RUN apk add --no-cache git
RUN git clone --depth 1 https://github.com/aeltai/k9s.git /tmp/k9s && \
    cd /tmp/k9s && go build -o rk9s . && mv /tmp/k9s/rk9s /rk9s

FROM alpine:latest
# gcompat: lets glibc-built krew plugins (e.g. browse-pvc) run on Alpine/musl
RUN apk add --no-cache curl git bash bash-completion runc openssh-client gcompat

# Zellij (terminal multiplexer)
RUN apk add --no-cache --repository https://dl-cdn.alpinelinux.org/alpine/edge/community zellij || \
    (cd /tmp && ZJ_VER=$(curl -sL https://api.github.com/repos/zellij-org/zellij/releases/latest | grep tag_name | cut -d'"' -f4) && \
     curl -fsSL "https://github.com/zellij-org/zellij/releases/download/${ZJ_VER}/zellij-${ZJ_VER}-x86_64-unknown-linux-musl.tar.gz" | tar xz && \
     mv zellij /usr/local/bin/ && chmod +x /usr/local/bin/zellij)

# crictl (container runtime CLI)
RUN cd /tmp && \
    CRICTL_VER=$(curl -sL https://api.github.com/repos/kubernetes-sigs/cri-tools/releases/latest | grep tag_name | cut -d'"' -f4) && \
    curl -fsSL "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VER}/crictl-${CRICTL_VER}-linux-amd64.tar.gz" | tar xz -C /usr/local/bin && \
    rm -rf /tmp/* 2>/dev/null || true

# etcdctl
RUN cd /tmp && \
    ETCD_VER=$(curl -sL https://api.github.com/repos/etcd-io/etcd/releases/latest | grep tag_name | cut -d'"' -f4) && \
    curl -fsSL "https://github.com/etcd-io/etcd/releases/download/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz" | tar xz && \
    mv etcd-${ETCD_VER}-linux-amd64/etcdctl /usr/local/bin/ && chmod +x /usr/local/bin/etcdctl && \
    rm -rf /tmp/etcd-* 2>/dev/null || true

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -sL https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && chmod +x kubectl && mv kubectl /usr/local/bin/

COPY --from=rk9s-builder /rk9s /usr/local/bin/rk9s
RUN chmod +x /usr/local/bin/rk9s

# Install GitHub CLI (gh) — issues, PRs, workflow runs, logs
RUN cd /tmp && \
    GH_VER=$(curl -sL https://api.github.com/repos/cli/cli/releases/latest | grep tag_name | cut -d'"' -f4) && \
    curl -fsSL "https://github.com/cli/cli/releases/download/${GH_VER}/gh_${GH_VER#v}_linux_amd64.tar.gz" | tar xz && \
    mv gh_*/bin/gh /usr/local/bin/ && chmod +x /usr/local/bin/gh && \
    rm -rf /tmp/gh_* 2>/dev/null || true

# Install krew to /opt/krew (non-root compatible)
RUN mkdir -p /opt/krew && \
    set -x && \
    OS="$(uname | tr '[:upper:]' '[:lower:]')" && \
    ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')" && \
    KREW="krew-${OS}_${ARCH}" && \
    curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" && \
    tar zxvf "${KREW}.tar.gz" && \
    KREW_ROOT=/opt/krew ./"${KREW}" install krew && \
    rm -f "${KREW}.tar.gz" "${KREW}"

ENV KREW_ROOT=/opt/krew
ENV PATH="/opt/krew/bin:${PATH}"

RUN kubectl krew update

# Keep a read-only copy so entrypoint can restore into volume when PVC overwrites /opt/krew (no runtime installer = no permission issues)
RUN cp -a /opt/krew /usr/local/share/krew-default

# Writable dirs and user for non-root run (UID 1000); entrypoint uses CONFIG_DIR=/opt/krew-workstation and KREW_ROOT=/opt/krew
RUN adduser -D -u 1000 krewuser && mkdir -p /opt/krew-workstation && chown -R 1000:1000 /opt/krew /opt/krew-workstation

COPY --from=builder /app/krew-manager /usr/local/bin/krew-manager
RUN chmod 755 /usr/local/bin/krew-manager
COPY kubectl-wrapper.sh /usr/local/bin/kubectl-wrapper
RUN chmod +x /usr/local/bin/kubectl-wrapper
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -f http://localhost:3000/health || exit 1

ENTRYPOINT ["/entrypoint.sh"]
