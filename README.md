# outline-bot


```commandline
go mod init outline-bot
go install .

go build 
./outline-bot

docker build . -t outline-bot:latest
docker run --env-file .env outline-bot

```