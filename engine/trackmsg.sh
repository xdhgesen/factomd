#/bin/sh
pattern=$1
shift
echo "grep -E \"$pattern\" $@ | awk -f debug/msgOrder.awk | sort -n | grep -E \"$pattern\" --color='always' | less -R"
grep -E \"$pattern\" $@
