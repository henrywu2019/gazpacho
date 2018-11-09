# Gazpacho

Example golang 1.11 compatible monorepo. Microservices are located in root directory, shared code lives under `g/`

```
.
├── g
│   ├── cfg
│   │   ├── cfg.go
│   │   ├── cfg_test.go
│   │   └── testdata
│   └── lib.go
├── go.mod
├── go.sum
├── out
├── project1-serviceA
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── project1-serviceA
├── project2
│   └── hello.php
└── README.md

5 directories, 12 files
```
