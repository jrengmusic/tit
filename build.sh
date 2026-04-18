#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

detect_cpu_count() {
    case "$(uname -s)" in
        Darwin)        sysctl -n hw.logicalcpu ;;
        Linux)         nproc ;;
        MINGW*|MSYS*)  nproc ;;
        *)             echo 4 ;;
    esac
}

echo "Cleaning..."
rm -rf Builds/Ninja

echo "Configuring..."
cmake -S . -B Builds/Ninja -G Ninja -DCMAKE_BUILD_TYPE=Release

echo "Building..."
cmake --build Builds/Ninja -- -j"$(detect_cpu_count)"

echo "Build succeeded."
echo "Binary: Builds/Ninja/titc_artefacts/Release/titc"
