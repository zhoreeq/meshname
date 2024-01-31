#!/bin/sh

die () {
    >&2 echo "$1"
    exit 1
}


(
    # Remove previous package construction
    rm -rf package || die "Failed to remove old packages"
)

(
    # Remove already built programs, build and test what you have built
    cd ".." || die "Failed to change directory"
    GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw" \
      make clean all test || die "Failed to build and test project"
)

# Install all necessary files in the package structure
install -Dm755 "../meshnamed" "package/usr/bin/meshnamed" || die "Failed to copy meshnamed"
install -Dm644 "../meshnamed.service" "package/usr/lib/systemd/system/meshnamed.service" || die "Failed to copy meshnamed service"
install -Dm644 "control.template" "package/DEBIAN/control" || die "Failed to copy control template"
install -Dm644 "copyright" "package/usr/share/doc/meshname/copyright" || die "Failed to copy copyright file"
install -Dm644 "../protocol.md" "package/usr/share/doc/meshname/protocol.md"

# Fix the path in the systemd-service file
sed -i "s|/usr/local/bin/meshname|/usr/bin/meshname|g" "package/usr/lib/systemd/system/meshnamed.service" || die "Failed to patch service file"

# Set the current architecture for the package
ARCH="$(dpkg --print-architecture)"
[ "$ARCH" ] || die "Failed to get architecture string"
sed "s/%ARCHITECTURE%/$ARCH/" -i package/DEBIAN/control || die "Failed to replace architecture in control template"

# Build the actual package
echo "Building package..."
dpkg-deb --root-owner-group --build package meshname.deb || die "Failed to build package"

echo "Package $(dpkg-deb --show meshname.deb | sed "s/\t/ /") successful built!"
