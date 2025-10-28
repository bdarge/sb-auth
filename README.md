## sb-auth

sb-auth is as an authentication app where calls to others apps gets authenticated by it.

### ecdsa keys

Used to sign tokens
```
openssl ecparam -name prime256v1 -genkey -noout -out priv_key.pem
openssl pkey -in priv_key.pem -pubout -out pub_key.pem
```