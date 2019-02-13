# roi

[![Build Status](https://travis-ci.com/studio2l/roi.svg?branch=master)](https://travis-ci.com/studio2l/roi)
[![Code Coverage](https://codecov.io/gh/studio2l/roi/branch/master/graph/badge.svg)](https://codecov.io/gh/studio2l/roi)
[![Go Report Card](https://goreportcard.com/badge/github.com/studio2l/roi)](https://goreportcard.com/report/github.com/studio2l/roi)


roi는 2L의 새 파이프라인 서버이자, 첫번째 오픈소스 프로젝트입니다!


## 주의

roi는 아직 디자인이 끝나지 않았으며, 구현의 초기 단계입니다.

즉, 언제든 모든 것이 바뀔수 있습니다.


## 설치

roi는 go1.11의 모듈 시스템을 사용합니다. go1.11 이상을 설치해 주세요.

.bashrc에 다음 줄을 추가해 모듈 사용을 켜주세요.

```
export GO111MODULE=on
```

### DB

roi는 [cockroach db](https://cockroachlabs.com) 를 사용합니다.

cockroach db는 postgresql의 호환 문법을 사용하며, 쉽게 스케일을 키울수 있는 db입니다.

cockroach db는 바이너리 파일로 배포하기 때문에 쉽게 설치하실 수 있습니다.

다음은 cockroachdb 홈페이지에 나와있는 설치법입니다.

v2.1.2는 현재 roi가 사용하는 버전입니다. 또는 [이 곳](https://www.cockroachlabs.com/docs/stable/install-cockroachdb.html)에서 최신 버전을 다운로드 받으실 수 있습니다.

```
wget -qO- https://binaries.cockroachdb.com/cockroach-v2.1.2.linux-amd64.tgz | tar  xvz
cp -i cockroach-v2.1.2.linux-amd64/cockroach /usr/local/bin
```

다운로드 받은 후 원하는 곳에서 실행하시면 그 아래에 cockroach-data 디렉토리가 생성되며 실행됩니다.

## 실행

```
# DB 실행
cd ~ # 또는 원하는 실행 장소에서
cockroach start --insecure &

# Test DB 추가
cd /roi/cmd/roishot/
go build
./roishot ./testdata/test.xlsx

# 서버 실행
export GO111MODULE=on # 혹시 빠뜨렸을 때에 대비해
git clone https://github.com/studio2l/roi
cd roi/cmd/roi
go build
./roi -init
sudo ./roi
```
