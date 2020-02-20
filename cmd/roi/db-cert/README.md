### DB 인증서를 위한 디렉토리

아래는 필요한 인증서를 생성하기 위한 명령어 입니다.

```
# 이 디렉토리에서
cockroach cert create-ca --certs-dir=. --ca-key=ca.key
cockroach cert create-node --certs-dir=. --ca-key=ca.key localhost
cockroach cert create-client --certs-dir=. --ca-key=ca.key root
```
