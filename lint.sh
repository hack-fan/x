set -e

lint() {
    (
        cd $1 || exit 1
        golangci-lint run
    )
}

pkgs="rdb xdb xecho xerr xlog xmp xobj xtype xpay"

for pkg in $pkgs; do
    lint "$pkg"
done
