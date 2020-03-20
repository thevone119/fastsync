#export GOARCH=amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64
export GOPATH=/home/ccb/gowork/fastsync
echo $GOPATH
go build testfsnotify.go
go build Server.go


#ps -ef | grep /FastSyncServer | awk '{print $2}' | xargs kill -9
#kill -9 `pidof FastSyncServer`
cd /home/ap/webapp/FastSync
nohup /home/ap/webapp/FastSync/FastSyncServer >/dev/null 2>&1&
nohup ./FastSyncServer >/dev/null 2>&1&
nohup /home/ap/ccb/FastSyncClient/FastSyncDir -p / -t 0 -c 4 >/dev/null 2>&1&

crontab -e
@reboot /home/ap/webapp/FastSync/.sh



nohup inotifywait -mrq --timefmt '%d/%m/%y/%H:%M' --format '%T %w %f' -e modify,delete,create,attrib /home/ccb  > inotifywait.log 2>&1 &