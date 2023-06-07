cd ~/nincomp/inc
go build -v -ldflags "-s -w"
mv --force --update --verbose inc ~/Desktop
cd ~/nincomp/com
go build -v -ldflags "-s -w"
mv --force --update --verbose com ~/Desktop