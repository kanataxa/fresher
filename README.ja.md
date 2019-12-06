# fresher

`fresher` はファイルやディレクトリを監視して監視対象のファイル群に変更があった場合に、Goアプリケーションを再起動(go run)させるCLIツールです。

このツールは、https://github.com/gravityblast/freshにある必要最低限の機能を実装し、またより効率よく監視していくためにいくつかの機能を追加したライブラリとなります。

## 機能
fresher は以下の機能をサポートしています。

1. 指定したディレクトリ群の変更を監視する
2. 指定したディレクトリ群の変更を無視する
3. 指定したファイル端子の変更を監視する
4. 監視のインターバルを指定する
5. テストファイル `*_test.go` を無視する

### 今後実装予定(仮)の機能
1. 正規表現によるファイル/ディレクトリの指定
2. ログ

## Install
go get github.com/kanataxa/fresher

## Usage
fresher -c your_config.yaml

## 設定ファイル

```yaml
build:
  dir: testdata
  file: a.go
paths:
  - pkg
exclude:
  - vendor
  - tmp
  - .git
extensions:
  - go
  - html
ignore_test: true
interval: 2
```

## License
MIT

## 備考
現在開発中のツールのため、動作の保証はできません。もし不具合などあれば、Issueに記述ください。