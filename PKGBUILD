pkgname=snesinfo-git
_pkgname=snesinfo
pkgver=latest.r0.g368459f
pkgrel=1
arch=('i686' 'x86_64')
url='https://github.com/Mat1az/snesinfo'
source=('git+https://github.com/Mat1az/snesinfo.git')
depends=()
makedepends=('go')
sha1sums=('SKIP')

pkgver() {
  cd "$srcdir/$_pkgname"
  ( set -o pipefail
    git describe --long --tags 2>/dev/null | sed 's/\([^-]*-g\)/r\1/;s/-/./g' ||
    printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
  )
}

build(){
  cd "$srcdir/$_pkgname"
  GO111MODULE=on go build -o "$srcdir/bin/snesinfo"
}

package() {
  cd "$srcdir/bin"
  install -Dm755 'snesinfo' "$pkgdir/usr/bin/snesinfo"
}
