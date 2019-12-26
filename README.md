# fresher
※ Click [here](https://github.com/kanataxa/fresher/blob/master/README.ja.md) for the Japanese version


`fresher` is a hot reload tool that restarts the Go application when the group of monitored files changes.

This tool is a library that implements the minimum required functions at https://github.com/gravityblast/fresh and add some functions to monitor more efficiently and efficiently.

Developing for `better-fresh`.

# Features
`fresher` supports the following features.

- Features of https://github.com/gravityblast/fresh
- Flexible setting of monitoring file
- Execute Command before and after build
- Support to set env and args for build
- Build with local and run with docker, hot reload faster with docker

# Install
```bash
go get github.com/kanataxa/fresher/cmd/fresher
```
# Quick Start
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
## Where is the example？
See here for a simple example in `_example`.

## Want to monitor or ignore files more closely
you can determine the files to be monitored in more detail by making two settings
1. Set `path` finely
2. Set `exclude`

Basically, the file to be monitored is determined from the configuration file according to the following rules.

1. Starting from the directory described in path, recursively determine whether it is a file to be monitored

2. If include is specified, only directories that are recursively viewed and files to be monitored should match those specified in include. If they match files set in exclude, Exclude from monitoring

3. include / exclude does not distinguish between files and directories, and uses filepath.Match to monitor / exclude matches

If you want to ignore everything, exclude it with `exclude`.
If you want to monitor / ignore specific files and directories in that directory, specify `include / exclude` in` path`.

The following is an example.

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

## Want to build a high-speed hot reload environment using Docker For Mac

### In case of CGO_ENABLED=0
Include the following settings in the `build` field, launch a` docker` container and execute `fresher` on the host side.
`GOOS` and` GOARCH` need to be set when the host and container OS are different.

`fresher` cross-compiles and runs the binary in` docker`.

``` yaml
build:
  host:
    docker: container_name
  env:
    GOOS: linux
    GOARCH: amd64
```

### In case of CGO_ENABLED=1

If you want to cross-compile binaries containing `cgo`, you need to prepare.

1. Prepare `c / c ++` compiler for cross compilation

2. If the c source used for cgo uses an external library, compile with the compiler prepared in step 1 and prepare it as a static library

3. Make the following settings in the configuration file

```yaml
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


# Bug reports and requests
Please create `Issue` in English or Japanese.

# License
MIT
