echo 'mode: collection' > coverage.txt
tail -q -n +2 all.cover.out >> coverage.txt
tail -q -n +2 impl.cover.out >> coverage.txt