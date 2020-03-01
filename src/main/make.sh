#export GOARCH=amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64
export GOPATH=/home/ccb/gowork/fastsync
echo $GOPATH
go build testfsnotify.go
go build Server.go


nohup ./FastSyncServer > FastSyncServer.log 2>&1 &
nohup ./testfsnotify > testfsnotify.log 2>&1 &



nohup inotifywait -mrq --timefmt '%d/%m/%y/%H:%M' --format '%T %w %f' -e modify,delete,create,attrib /home/ccb  > inotifywait.log 2>&1 &