dev::
	go build -o quicd

all:: quicd_illumos

quicd_illumos::
	env GOOS=solaris GOARCH=amd64 go build -o $@

.PHONY: ssl-keys
ssl-keys::
	openssl req -subj '/CN=quicd.example.com/O=Joyent Inc./C=US' -new -newkey rsa:2048 -days 3650 -nodes -x509 -sha256 -keyout quicd.key -out quicd.pem

run::
	./quicd -circonus-api-key=`cat .token.key`
