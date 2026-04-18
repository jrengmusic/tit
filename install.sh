#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

detect_os() {
    case "$(uname -s)" in
        Darwin) echo "macos" ;;
        Linux)  echo "linux" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) echo "unknown" ;;
    esac
}

OS="$(detect_os)"

echo "Cleaning..."
rm -rf Builds

echo "Building..."
case "$OS" in
    windows)
        cmd.exe //c build.bat clean Release
        ;;
    macos)
        cmake -S . -B Builds/Ninja -G Ninja -DCMAKE_BUILD_TYPE=Release
        cmake --build Builds/Ninja -- -j"$(sysctl -n hw.logicalcpu)"
        ;;
    linux)
        cmake -S . -B Builds/Ninja -G Ninja -DCMAKE_BUILD_TYPE=Release
        cmake --build Builds/Ninja -- -j"$(nproc)"
        ;;
    *)
        echo "Unsupported OS: $(uname -s)"
        exit 1
        ;;
esac

echo "Installing..."
case "$OS" in
    windows)
        ARTIFACT="Builds/Ninja/titc_artefacts/Release/titc.exe"
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
        cp "$ARTIFACT" "$INSTALL_DIR/titc.exe"
        ;;
    macos)
        ARTIFACT="Builds/Ninja/titc_artefacts/Release/titc"
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
        cp "$ARTIFACT" "$INSTALL_DIR/titc"
        ;;
    linux)
        ARTIFACT="Builds/Ninja/titc_artefacts/Release/titc"
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
        cp "$ARTIFACT" "$INSTALL_DIR/titc"
        ;;
esac

echo "Done."
