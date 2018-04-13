#/bin/sh
#grep -E \"\" $@ | awk -f debug/msgOrder.awk | sort -n | less -R
echo $@
grep -E . $@  | awk -f debug/msgOrder.awk | sort -n | less -R

