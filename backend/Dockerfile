FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o krew-manager .

FROM alpine:latest
RUN apk add --no-cache curl git bash bash-completion runc openssh-client make

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

# Helm
RUN cd /tmp && \
    HELM_VER=$(curl -sL https://api.github.com/repos/helm/helm/releases/latest | grep '"tag_name"' | head -1 | cut -d'"' -f4) && \
    curl -fsSL "https://get.helm.sh/helm-${HELM_VER}-linux-amd64.tar.gz" | tar xz && \
    mv linux-amd64/helm /usr/local/bin/ && chmod +x /usr/local/bin/helm && \
    rm -rf /tmp/linux-amd64

# Rancher CLI
RUN cd /tmp && \
    RANCHER_VER=$(curl -sL https://api.github.com/repos/rancher/cli/releases/latest | grep '"tag_name"' | head -1 | cut -d'"' -f4) && \
    curl -fsSL "https://github.com/rancher/cli/releases/download/${RANCHER_VER}/rancher-linux-amd64-${RANCHER_VER}.tar.gz" | tar xz && \
    mv rancher-${RANCHER_VER}/rancher /usr/local/bin/rancher && chmod +x /usr/local/bin/rancher && \
    rm -rf /tmp/rancher-* 2>/dev/null || true

# Fleet CLI (single binary, not tarball)
RUN cd /tmp && \
    FLEET_VER=$(curl -sL https://api.github.com/repos/rancher/fleet/releases/latest | grep '"tag_name"' | head -1 | cut -d'"' -f4) && \
    curl -fsSL "https://github.com/rancher/fleet/releases/download/${FLEET_VER}/fleet-linux-amd64" \
      -o /usr/local/bin/fleet && chmod +x /usr/local/bin/fleet

# kwctl (Kubewarden) — shipped as zip
RUN apk add --no-cache unzip && cd /tmp && \
    KWCTL_VER=$(curl -sL https://api.github.com/repos/kubewarden/kwctl/releases/latest | grep '"tag_name"' | head -1 | cut -d'"' -f4) && \
    curl -fsSL "https://github.com/kubewarden/kwctl/releases/download/${KWCTL_VER}/kwctl-linux-x86_64.zip" -o kwctl.zip && \
    unzip -j kwctl.zip kwctl-linux-x86_64 -d /usr/local/bin/ && \
    mv /usr/local/bin/kwctl-linux-x86_64 /usr/local/bin/kwctl && chmod +x /usr/local/bin/kwctl && \
    rm -f kwctl.zip

# virtctl (KubeVirt)
RUN cd /tmp && \
    VIRT_VER=$(curl -sL https://api.github.com/repos/kubevirt/kubevirt/releases/latest | grep '"tag_name"' | head -1 | cut -d'"' -f4) && \
    curl -fsSL "https://github.com/kubevirt/kubevirt/releases/download/${VIRT_VER}/virtctl-${VIRT_VER}-linux-amd64" \
      -o /usr/local/bin/virtctl && chmod +x /usr/local/bin/virtctl

# longhornctl
RUN cd /tmp && \
    LH_VER=$(curl -sL https://api.github.com/repos/longhorn/cli/releases/latest | grep '"tag_name"' | head -1 | cut -d'"' -f4) && \
    curl -fsSL "https://github.com/longhorn/cli/releases/download/${LH_VER}/longhornctl-linux-amd64" \
      -o /usr/local/bin/longhornctl && chmod +x /usr/local/bin/longhornctl

# Install k9s (standalone binary — more reliable than krew in restricted networks)
RUN cd /tmp && \
    K9S_VER=$(curl -sL https://api.github.com/repos/derailed/k9s/releases/latest | grep tag_name | cut -d'"' -f4) && \
    curl -fsSL "https://github.com/derailed/k9s/releases/download/${K9S_VER}/k9s_Linux_amd64.tar.gz" | tar xz && \
    mv k9s /usr/local/bin/ && chmod +x /usr/local/bin/k9s && \
    rm -rf /tmp/* 2>/dev/null || true

# Install krew
RUN set -x && \
    OS="$(uname | tr '[:upper:]' '[:lower:]')" && \
    ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')" && \
    KREW="krew-${OS}_${ARCH}" && \
    curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/${KREW}.tar.gz" && \
    tar zxvf "${KREW}.tar.gz" && \
    ./"${KREW}" install krew && \
    rm -f "${KREW}.tar.gz" "${KREW}"

ENV KREW_ROOT=/root/.krew
ENV PATH="/root/.krew/bin:${PATH}"

RUN kubectl krew update

COPY --from=builder /app/krew-manager /usr/local/bin/krew-manager
COPY --from=builder /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/root/go
ENV GOMODCACHE=/root/go/pkg/mod

COPY entrypoint.sh install-local-plugins.sh install-local-binaries.sh /
RUN chmod +x /entrypoint.sh /install-local-plugins.sh /install-local-binaries.sh

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -f http://localhost:3000/health || exit 1

ENTRYPOINT ["/entrypoint.sh"]
