# fresher

`fresher` はファイルやディレクトリを監視して監視対象のファイル群に変更があった場合に、Goアプリケーションを再起動(go run)させるCLIツールです。

このツールは、https://github.com/gravityblast/fresh にある必要最低限の機能を実装し、またより効率よく監視していくためにいくつかの機能を追加したライブラリとなります。

`better-fresh` を目指し開発をしています。

## 機能
`fresher` は以下の機能をサポートしています。

1. 指定したディレクトリ群の変更を監視する
2. 指定したファイル端子の変更を監視する
3. 指定したディレクトリ群の変更を無視する
4. 指定したパターンに一致するファイルを無視する
5. 監視のインターバルを指定する

### 今後実装予定(仮)の機能
1. ログ

## Install
```bash
go get github.com/kanataxa/fresher
```
## Usage
```bash
fresher -c your_config.yaml
```
## 設定ファイル

```yaml
command:
  name: go
  args: 
    - run
    - ./main.go
path:
  - name: model
  - name: pkg
    dir:
      - name: utils
    exclude:
      - const.go
exclude:
  dir:
    - vendor
  file:
    - '*_test.go'
extension:
  - go
interval: 2
```

## License
MIT

## 備考
現在開発中のツールのため、動作の保証はできません。もし不具合などあれば、Issueに記述ください。