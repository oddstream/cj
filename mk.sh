cd inc
go build -v -ldflags "-s -w"
mv --force --update --verbose inc ~/Desktop