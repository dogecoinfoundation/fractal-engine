for file in docs/mermaid/**/*.mmd; do
  mmdc -i "$file" -o "${file%.mmd}.png" -s 2
done
