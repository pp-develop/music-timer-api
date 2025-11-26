FROM golang:1.24.1

# Install Go development tools
RUN go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install github.com/air-verse/air@latest

# Install Python and required dependencies
RUN apt-get update && apt-get install -y \
    python3 \
    python3-pip \
    python3-venv \
    && rm -rf /var/lib/apt/lists/*

# Create symbolic link for python command
RUN ln -s /usr/bin/python3 /usr/bin/python

# Copy requirements.txt and install Python dependencies
WORKDIR /tmp
COPY requirements.txt .
RUN pip3 install --no-cache-dir -r requirements.txt --break-system-packages

# Verify installations
RUN python --version && \
    pip3 --version && \
    python -c "import ytmusicapi; print(f'ytmusicapi {ytmusicapi.__version__} installed')"

# Set working directory back to default
WORKDIR /go/src

CMD ["air"]
