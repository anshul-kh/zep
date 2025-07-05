#!/usr/bin/env bash

set -e

echo "---------------------------------------------"
echo "Installing Go and Make for your Linux distro."
echo "---------------------------------------------"

# Detect distro
if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO=$ID
else
    echo "Cannot detect Linux distribution."
    exit 1
fi

echo "Detected distro: $DISTRO"

# Function to install packages
install_debian() {
    sudo apt update
    sudo apt install -y golang-go make
}

install_fedora() {
    sudo dnf install -y golang make
}

install_centos() {
    sudo yum install -y golang make
}

install_arch() {
    sudo pacman -Sy --noconfirm go make
}

case "$DISTRO" in
    ubuntu|debian)
        install_debian
        ;;
    fedora)
        install_fedora
        ;;
    centos|rhel)
        install_centos
        ;;
    arch|endeavouros)
        install_arch
        ;;
    *)
        echo "Unsupported distro: $DISTRO"
        exit 1
        ;;
esac

echo "---------------------------------------------"
echo "Go and make installed successfully!"
echo "---------------------------------------------"
