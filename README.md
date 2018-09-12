# roi

roi는 2L의 새 파이프라인 서버이자, 첫번째 오픈소스 프로젝트입니다!


## 주의

roi는 아직 디자인이 끝나지 않았으며, 구현의 초기 단계입니다.

즉, 언제든 모든 것이 바뀔수 있습니다.


## 설치

설치는 일반적인 go 프로그램의 설치와 같은 방법으로 합니다.

```
go get github.com/studio2l/roi
cd $GOPATH/src/github.com/studio2l/roi
go install ./...
```

### DB

저희는 [cockroach db](https://cockroachlabs.com) 사용을 권장합니다.

cockroach db는 postgres와 호환되며, 쉽게 스케일을 키울수 있는 db입니다.

또는 postgres를 사용하셔도 됩니다.

cockroach db는 바이너리 파일로 배포하기 때문에 쉽게 설치하실 수 있습니다.

다운로드 받은 후 원하는 곳에서 실행하시면

그 아래에 cockroach-data 디렉토리가 생성되며 실행됩니다.

## 실행

```
# DB 실행 및 최초 셑업
cd ~ # 또는 원하는 실행 장소에서
cockroach start --insecure
cockroach sql --insecure
> CREATE USER maxroach;
> CREATE DATABASE roi;
> GRANT ALL ON roi TO maxroach;
> \q

# 테스트 데이터 추가
cd $GOPATH/github.com/studio2l/roi/cmd/roishot
go install
roishot testdata/test.xlsx

# 자가 https 인증서 추가
cd $GOPATH/github.com/studio2l/roi/cmd/roi/cert
sh generate-self-signed-cert.sh

# 서버 실행
cd $GOPATH/github.com/studio2l/roi/cmd/roi
go install
roi
```
