# epoll chat
system callのepollを直接使って、chat serverを作るという試み  

# 環境
go: 1.15.8  
os: ubuntu20.04  
※ linuxのみ動作します(MacとかFreeBSDは動作しない)  

# server run
```go
go run main.go
```

# chatの仕方
```ssh

// 下記のコマンドで複数のterminalでログイン
telnet localhost 2000  

// 任意のコメントを書き込んで、Enterを押すと別のClientで表示される

// telnetのlogout
// ctrl + ]
// その後
// qを入力して、Enter


```

# 参考

https://keens.github.io/blog/2021/02/01/epolldetsukuruchattosa_ba/  
https://gist.github.com/tevino/3a4f4ec4ea9d0ca66d4f

# TODO 
結構ぐちゃぐちゃなので整理する  
kqueue/keventでも実装してみる  
もうちょっとsyscallの使い方を調べる  

