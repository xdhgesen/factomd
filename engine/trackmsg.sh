#/bin/sh
grep -E "$1" $2 | awk -f debug/msgOrder.awk | sort -n | grep -E "$1" --color='always' | less -R
