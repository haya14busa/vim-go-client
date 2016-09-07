#!/bin/sh
# ref: https://github.com/vim-jp/vital.vim/blob/bff0d8c58c1fb6ab9e4a9fc0c672368502f10d88/scripts/install-vim.sh
set -e

git clone --depth 1 https://github.com/vim/vim /tmp/vim
cd /tmp/vim
./configure --prefix="$HOME/vim" --with-features=huge --enable-fail-if-missing
make -j2
make install
