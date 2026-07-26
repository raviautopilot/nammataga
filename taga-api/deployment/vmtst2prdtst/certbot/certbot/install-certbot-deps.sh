#!/bin/bash

# Python dependencies for Certbot

set -e  # Exit on error

echo "Installing Python dependencies for Certbot..."

# Install required system packages
sudo apt update
sudo apt install -y \
    python3 \
    python3-dev \
    python3-venv \
    python3-pip \
    libffi-dev \
    libssl-dev \
    cargo

# Install Python packages needed by Certbot
sudo pip3 install --upgrade pip
sudo pip3 install --upgrade \
    setuptools \
    wheel \
    cryptography \
    certbot

echo "Python dependencies for Certbot installed successfully!"
