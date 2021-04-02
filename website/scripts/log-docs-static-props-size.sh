cd .next/server/pages/docs
find . -name "*.json" -print0 | while read -d $'\0' file; do
    size_kb=$(du -k "$file" | cut -f1)
    basename=${file##*/}
    printf "%4s kB | $basename\n" "$size_kb"
done
