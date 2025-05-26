openssl req -x509 -new -nodes -config localhost.cnf -extensions v3_req \
  -days 365 -keyout key.pem -out cert.pem
