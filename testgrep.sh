# grep --fixed-strings --recursive --ignore-case --files-with-matches -I dog /home/gilbert/.cj/Default
# grep --fixed-strings --recursive --ignore-case --files-with-matches -I 'a few years' /home/gilbert/.cj/Default
# grep --extended-regexp --recursive --ignore-case --only-matching --no-filename -I '#[[:alnum:]]+' /home/gilbert/.cj/Default
grep --extended-regexp --recursive --ignore-case --only-matching --no-filename -I '#\w+' /home/gilbert/.cj/Default