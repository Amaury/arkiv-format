#!/bin/sh

# ###################################################################
# # Test suite for Arkiv format                                     #
# #-----------------------------------------------------------------#
# # Copyright © 2025, Amaury Bouchard <amaury@amaury.net>           #
# # Published under the terms of the MIT license.                   #
# # https://opensource.org/license/mit                              #
# ###################################################################

# ########## UTILITY FUNCTIONS ##########
# Print an error message and exit.
fail() {
	echo "$(tput setab 1)$*$(tput sgr0)" >&2
	exit 1
}
# Print a success message.
success() {
	echo "$(tput setab 2)$*$(tput sgr0)"
}

# ########## INIT ##########
PATH=../shell/:$PATH
export ARKIV_PASS="$(head -c 10 /dev/urandom | base64)"

# ########## TEST 1: FILES ##########
# archive creation
mkdir res-01
cd src-01
if ! ../../shell/arkiv-build ../a.arkiv a.txt z.txt; then
	echo "'$PATH'"
	cd - > /dev/null
	rm -rf ./a.arkiv ./res-01
	fail "TEST 1: arkiv-build"
fi
cd - > /dev/null
# list archive content
if [ "$(arkiv-tree a.arkiv | grep "a.txt")" = "" ] ||
   [ "$(arkiv-tree a.arkiv | grep "z.txt")" = "" ]; then
	rm -rf ./a.arkiv ./res-01
	fail "TEST 1: arkiv-tree"
fi
# extract the whole archive in the current directory
if ! arkiv-extract a.arkiv ||
   ! ls ./a.txt > /dev/null 2>&1 ||
   ! ls ./z.txt > /dev/null 2>&1; then
	rm -rf ./a.arkiv ./a.txt ./z.txt ./res-01
	fail "TEST 1: arkiv-extract (same directory)"
fi
rm -f ./a.txt ./z.txt
# extract the whole archive in a sub-directory
if ! arkiv-extract a.arkiv res-01 ||
   [ "$(cat res-01/a.txt 2> /dev/null)" != "abcde" ] ||
   [ "$(cat res-01/z.txt 2> /dev/null)" != "zyxwv" ]; then
	rm -rf ./a.arkiv ./res-01
	fail "TEST 1: arkiv-extract (subdir)"
fi
rm -rf ./res-01
# extract one file
mkdir res-01
if ! arkiv-extract a.arkiv res-01 z.txt ||
   [ "$(cat res-01/z.txt 2> /dev/null)" != "zyxwv" ] ||
   [ "$(ls -l res-01/z.txt | grep "\-rwxr-xr-x" 2> /dev/null)" = "" ]; then
	rm -rf ./a.arkiv ./res-01
	fail "TEST 1: arkiv-extract (one file)"
fi
rm -rf ./a.arkiv ./res-01
success "TEST 1"

# ########## TEST 2: DIRECTORIES ##########
# archive creation
mkdir res-02
if ! arkiv-build a.arkiv src-02; then
	rm -rf ./a.arkiv ./res-02
	fail "TEST 2: arkiv-build"
fi
# list archive content
if [ "$(arkiv-tree a.arkiv | grep "src-02/sub2/sub3/z.txt")" = "" ]; then
	rm -rf ./a.arkiv ./res-02
	fail "TEST2: arkiv-tree"
fi
# extract the whole archive in a sub-directory
if ! arkiv-extract a.arkiv res-02 ||
   [ "$(cat "res-02/src-02/sub1/a.txt" 2> /dev/null)" != "abcde" ] ||
   [ "$(cat "res-02/src-02/sub2/sub3/z.txt" 2> /dev/null)" != "zyxwv" ]; then
	rm -rf ./a.arkiv ./res-02
	fail "TEST 2: arkiv-extract"
fi
rm -rf ./res-02
# extract one file
mkdir res-02
if ! arkiv-extract a.arkiv res-02 "src-02/sub2/sub3/z.txt" ||
   [ "$(cat "res-02/src-02/sub2/sub3/z.txt" 2> /dev/null)" != "zyxwv" ] ||
   [ "$(ls -l "res-02/src-02/sub2/sub3/z.txt" | grep "\-rwxr-xr-x" 2> /dev/null)" = "" ]; then
	rm -rf ./a.arkiv ./res-02
	fail "TEST 2: arkiv-extract (one file)"
fi
rm -rf ./a.arkiv ./res-02
success "TEST 2"

