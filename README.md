# roi

[![Build Status](https://travis-ci.com/studio2l/roi.svg?branch=master)](https://travis-ci.com/studio2l/roi)
[![Go Report Card](https://goreportcard.com/badge/github.com/studio2l/roi)](https://goreportcard.com/report/github.com/studio2l/roi)


roi는 2L의 새 파이프라인 서버이자, 첫번째 오픈소스 프로젝트입니다!


## 주의

roi는 아직 디자인이 끝나지 않았으며, 구현의 초기 단계입니다.

즉, 언제든 모든 것이 바뀔수 있습니다.


## 설치

roi는 go1.13 이상의 버전에서 컴파일 하는 것을 추천합니다.

아래에서는 `~/roi` 를 레포지터리 루트로 하여 설치 및 실행하는 방법을 다룹니다.

```
cd ~
git clone https://github.com/studio2l/roi
cd ~/roi/cmd/roi
go build
```

### DB

roi는 [cockroach db](https://cockroachlabs.com) 를 사용합니다.

cockroach db는 postgresql의 호환 문법을 사용하며, 쉽게 스케일을 키울수 있는 db입니다.

[이 곳](https://www.cockroachlabs.com/docs/stable/install-cockroachdb.html)에서 최신 버전을 다운로드 받으실 수 있습니다.

다운로드 받은 후 원하는 곳에서 실행하시면 그 아래에 cockroach-data 디렉토리가 생성되며 실행됩니다.

cockroach db를 실행하려면 우선 db 인증서를 생성해야 합니다.

```
cd ~/roi/cmd/roi
./init-db
```

## 실행

roi를 실행하기전 먼저 DB를 실행해야 합니다.

```
cd ~/roi/cmd/roi
./start-db
```

이제 새 터미널에서 roi를 실행합니다. 아래 예제에서는 -insecure 플래그를 써 http 프로토콜로
프로그램을 실행하지만, 도메인이 있는 서버에서는 네트워크간 전송되는 정보를 보호할 수 있는
https 프로토콜을 사용하는 것이 좋습니다.

```
# 서버 실행
cd ~/roi/cmd/roi
sudo ./roi -insecure
```

이제 http://localhost 페이지를 살펴보세요.

### 자가서명인증서 (Self-Signed Certificate) 생성

https 프로토콜을 사용하고 싶으나, 비용 또는 여타 문제로
인증서 구매/발급이 어려울 때 자가서명인증서를 사용하는 경우가 있습니다.

이 때 제가 추천하는 프로그램은 mkcert 입니다.

mkcert는 [여기](https://github.com/FiloSottile/mkcert)서 받을수 있습니다.

```
cd ~/roi/cmd/roi/cert
mkcert -install
mkcert -cert-file=cert.pem -key-file=key.pem localhost
```

### https 프로토콜 사용

-insecure 플래그를 제외하면 기본적으로 roi는 https 프로토콜을 사용합니다.

```
cd ~/roi/cmd/roi
sudo ./roi
```

### 환경변수

로이가 사용하는 환경변수는 다음과 같습니다.

```
ROI_ADDR: -addr 플래그를 지정하지 않았을때 서버가 바인딩하고, 클라이언트가 접근하는 주소입니다.
ROI_DB_ADDR: -db-addr 플래그를 지정하지 않았을때 서버가 사용하는 DB 주소입니다.
ROI_DB_HTTP_ADDR: start-db.sh가 사용하는 DB의 웹 서비스 주소입니다.
```
