pkgname=snesinfo-git
_pkgname=snesinfo
pkgver=0.0.1
pkgrel=1
pkgdesc='A Simple Tool for SNES ROM Info'
arch=('x86_64')
url="https://github.com/Mat1az/snesinfo"
license=('LGPL')
makedepends=('go')
source=("git+https://github.com/Mat1az/snesinfo.git")
md5sums=('SKIP')
options=('!debug')

prepare(){
  cd "$_pkgname"
  mkdir -p build/
}

build() {
  cd "$_pkgname"
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"
  go build -o build/
}

package() {
  cd "$_pkgname"
  install -Dm755 build/$_pkgname "$pkgdir"/usr/bin/$_pkgname
}