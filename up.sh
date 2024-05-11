
up() {
    (
        cd $1 || exit 1
        go get -u ./...
        go mod tidy
        go get ./...
    )
}

pkgs="rdb xdb xecho xerr xlog xmp xobj xtype xpay"

for pkg in $pkgs; do
    up "$pkg"
done
