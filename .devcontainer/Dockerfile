FROM mcr.microsoft.com/devcontainers/base:ubuntu

# Install Go
ARG GO_VERSION=1.21
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

# Set Go environment variables
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
ENV PATH=$PATH:$GOPATH/bin
ENV GO111MODULE=on

# Install additional development tools
RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    vim \
    nano \
    tree \
    htop \
    jq \
    && rm -rf /var/lib/apt/lists/*

# Install Go tools
RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install github.com/axw/gocov/gocov@latest && \
    go install github.com/AlekSi/gocov-xml@latest && \
    go install github.com/tebeka/go2xunit@latest

# Create workspace directory
WORKDIR /workspace

# Set up git configuration
RUN git config --global init.defaultBranch main

# Create a non-root user
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME \
    && apt-get update \
    && apt-get install -y sudo \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME

# Switch to non-root user
USER $USERNAME

# Set up shell
RUN echo 'export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin' >> ~/.bashrc && \
    echo 'export GOPATH=/go' >> ~/.bashrc && \
    echo 'export GO111MODULE=on' >> ~/.bashrc

# Default command
CMD ["/bin/bash"] 