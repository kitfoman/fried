ARG CUDA_VERSION=11.8.0
ARG UBUNTU_VERSION=20.04

FROM nvidia/cuda:${CUDA_VERSION}-devel-ubuntu${UBUNTU_VERSION} as base

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive

# Install common dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    ca-certificates \
    cmake \
    curl \
    git \
    libnuma-dev \
    libopenmpi-dev \
    openmpi-bin \
    python3 \
    python3-dev \
    python3-pip \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Create symbolic links for python
RUN ln -sf /usr/bin/python3 /usr/bin/python && \
    ln -sf /usr/bin/pip3 /usr/bin/pip

# Install PyTorch with CUDA support
ARG PYTORCH_VERSION=2.0.1
RUN pip install --no-cache-dir torch==${PYTORCH_VERSION}+cu$(echo ${CUDA_VERSION} | sed 's/\.//' | cut -c 1-3) \
    torchvision torchaudio \
    --extra-index-url https://download.pytorch.org/whl/cu$(echo ${CUDA_VERSION} | sed 's/\.//' | cut -c 1-3)

# Install DCGM diagnostic tools
RUN curl -s -L https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2004/x86_64/cuda-keyring_1.0-1_all.deb \
    -o cuda-keyring.deb \
    && dpkg -i cuda-keyring.deb \
    && apt-get update \
    && apt-get install -y --no-install-recommends datacenter-gpu-manager \
    && rm -rf /var/lib/apt/lists/* \
    && rm cuda-keyring.deb

# Install HPL (Linpack) dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    libopenblas-dev \
    && rm -rf /var/lib/apt/lists/*

# Download and build HPL benchmark
WORKDIR /opt
RUN wget -q http://www.netlib.org/benchmark/hpl/hpl-2.3.tar.gz \
    && tar -xzf hpl-2.3.tar.gz \
    && cd hpl-2.3 \
    && cp setup/Make.Linux_PII_CBLAS . \
    && sed -i 's|^TOPdir.*|TOPdir = $(HOME)/hpl-2.3|' Make.Linux_PII_CBLAS \
    && sed -i 's|^MPdir.*|MPdir = /usr|' Make.Linux_PII_CBLAS \
    && sed -i 's|^LAdir.*|LAdir = /usr|' Make.Linux_PII_CBLAS \
    && sed -i 's|^CC.*|CC = mpicc|' Make.Linux_PII_CBLAS \
    && sed -i 's|^LINKER.*|LINKER = mpicc|' Make.Linux_PII_CBLAS \
    && sed -i 's|^ARCH.*|ARCH = ar|' Make.Linux_PII_CBLAS \
    && sed -i 's|^HPL_OPTS.*|HPL_OPTS = -DHPL_DETAILED_TIMING -DHPL_PROGRESS_REPORT|' Make.Linux_PII_CBLAS \
    && sed -i 's|^LAlib.*|LAlib = -lopenblas|' Make.Linux_PII_CBLAS \
    && make arch=Linux_PII_CBLAS \
    && mkdir -p /opt/hpl-linux-x86_64/bin \
    && cp bin/Linux_PII_CBLAS/xhpl /opt/hpl-linux-x86_64/bin/ \
    && cp bin/Linux_PII_CBLAS/HPL.dat /opt/hpl-linux-x86_64/bin/ \
    && cd .. \
    && rm -rf hpl-2.3.tar.gz

# Create a script directory for custom PyTorch/CUDA scripts
WORKDIR /scripts
RUN mkdir -p /results

# Create entrypoint script to handle different run modes
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["all"]