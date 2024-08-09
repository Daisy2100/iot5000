# High Frequency Device Data Reader

高頻率讀取5000個設備後，將資料進行處理與轉換，最後存入資料庫。

## 使用到的框架和庫

- [Goroutines](https://tour.golang.org/concurrency/1)：Golang 原生的併發機制，輕量級執行緒管理。
- [Channels](https://tour.golang.org/concurrency/2)：用於 Goroutines 之間的通訊

## 環境要求

- [Go](https://golang.org/dl/) 1.16+

## install

```
go mod tidy
```

## run

```
go install github.com/gravityblast/fresh@latest
```

```
fresh
```

## build

```
go build main.go
```


for linux
```
GOOS=linux GOARCH=amd64 go build -o myapp-linux main.go
```


for macOS

```
GOOS=darwin GOARCH=amd64 go build -o myapp-darwin main.go
```

for win

```
GOOS=windows GOARCH=amd64 go build -o myapp.exe main.go
```

## execute

for linux and macOS

```
./myapp
```

for win

```
myapp.exe
```