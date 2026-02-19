FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o krew-manager .

FROM alpine:latest
RUN apk add --no-cache curl git bash bash-completion runc openssh-client

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
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -f http://localhost:3000/health || exit 1

ENTRYPOINT ["/entrypoint.sh"]
