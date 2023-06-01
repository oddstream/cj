cd ~/goldnotebook/inc
go build -v -ldflags "-s -w"
mv --force --update --verbose inc ~/Desktop
cd ~/goldnotebook/com
go build -v -ldflags "-s -w"
mv --force --update --verbose com ~/Desktop