# fresher

`fresher` はファイルやディレクトリを監視して監視対象のファイル群に変更があった場合に、Goアプリケーションを再起動させるホットリロードツールです。

このツールは、https://github.com/gravityblast/fresh にある必要最低限の機能を実装し、またより効率よく監視していくためにいくつかの機能を追加したライブラリとなります。

`better-fresh` を目指し開発をしています。

# 機能
`fresher` は以下の機能をサポートしています。

- https://github.com/gravityblast/fresh が持つ機能
- 監視ファイルの柔軟な設定
- ビルド前後でのCommandの実行
- build時のenvやargsのサポート
- localでビルドしdockerで実行でき、dockerを用いた環境でより高速にホットリロードできる

# インストール
```bash
go get github.com/kanataxa/fresher/cmd/fresher
```
# 使い方
## 1. Install CLI
```
go get github.com/kanataxa/fresher/cmd/fresher
```

## 2. Create YAML Config
```
touch your_config.yml
```

## 3. Edit YAML config
```yml
path:
  - .
```

## 4. Run fresher
```bash
fresher -c your_config.yml
```

# FAQ
## サンプルを見たい
`_example` に簡単なサンプルがあるので、こちらをご覧ください。

## ファイルをより細かく監視もしくは無視したい
fresher では2つの設定を行うことでより詳細に監視対象のファイルを決定できます
1. `path` を細かく設定する
2. `exclude` を設定する

基本的に以下のようなルールで設定ファイルから監視対象のファイルを決定します。
1. path に記述されたディレクトリを起点に再帰的に監視対象のファイルかどうかを判断する
2. include が指定されている場合は、再帰的に見るディレクトリや監視対象とするファイルを include で指定されたものと一致するもののみにする. もしくは exclude に設定されているファイルと一致した場合は監視対象から除外する
3. include/exclude はファイルかディレクトリかを区別せず filepath.Match を用いて一致したものを監視/除外する

もし、全てにおいて無視したいのであれば、 `exclude` で除外してください。
そのディレクトリ内で特定のファイルやディレクトリを監視/無視したいのであれば `path`の中の `include/exclude` を指定してください。

以下は一例です。
```
bash-3.2$ tree ./
./
├── fresher_config.yml
├── main.go
├── controller
│   ├── controller.go
│   └── const.go
├── model
│   ├── message.go
│   └── message_test.go
└── pkg
    ├── config.go
    ├── const.go
    ├── pkg.go
    └── utils
        ├── utils.go
        └── utils_test.go
```
このディレクトリの `fresher_config.yml` は、以下のようになっています

``` yaml
path:
  - .
  - name: controller
    exclude:
      - const.go
  - model
  - name: pkg
    include:
      - utils*
      - config.go
exclude:
  - vendor
  - '*_test.go'
```

## Docker For Macを用いて高速なホットリロード環境を構築したい
### CGO_ENABLED=0の場合
以下のような設定を `build` フィールドに含めて、`docker`コンテナを立ち上げホスト側で `fresher` を実行してください。
`GOOS` と `GOARCH` は、ホストとコンテナのOSが異なる場合に設定が必要です。
`fresher` はクロスコンパイルし、 `docker` 内でそのバイナリを実行します。
``` yaml
build:
  host:
    docker: container_name
  env:
    GOOS: linux
    GOARCH: amd64
```

### CGO_ENABLED=1の場合
`cgo` を含めたバイナリをクロスコンパイルしたい場合、いくつかの準備が必要になります。

1. クロスコンパイル用の `c/c++` コンパイラを用意する
2. cgoに使いたいcのソースが外部ライブラリを用いている場合、1で用意したコンパイラでコンパイルし、静的ライブラリとして用意する
3. 設定ファイルに以下の設定をする

```
build:
  host:
    docker: container
  env:
    GOOS: linux
    GOARCH: amd64
    CC: /your/cross-compile-cc-compiler/path
    CXX: /your/cross-compile-cxx-compiler/path
    CGO_ENABLED: 1
    CGO_LDFLAGS:  /your/static/lib/libz.a
    CGO_CXXFLAGS: -I/your/header/path/include
  arg:
    - --ldflags
    - '-linkmode external -extldflags -static'
```

### 備考
この機能は現在 `os/exec` を使っていますが、将来的には `Docker SDK` を使うことも考えています。

# バグ報告や要望など
`Issue` を英語もしくは日本語で立ててください。

# License
MIT
