package main

import (
	"epollchat/pkg/common/chk"
	"epollchat/pkg/common/jsonutil"
	"fmt"
	"log"
	"syscall"
)

const (
	Epollet = 1 << 31

	MaxEpollEvents = 32
)

// User TODO 複数thredからのアクセスを安全にする
type User struct {
	Fd int
}

func main() {
	fmt.Println("epoll chat")
	var eventList [MaxEpollEvents]syscall.EpollEvent // バッファの確保

	// 通信のためのSocketを作成する
	// AF_INETでIPv4企画の通信を採用する
	// O_NONBLOCK:ファイル非停止モードでOpenする
	// SOCK_STREAM: 順序保証, 信頼せありの双方向のbyte streamを提供
	// | ビット演算子 orで複数のフラグをここで立てているみたい
	fd, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	chk.SE(err, "Socketの作成に失敗")

	// ファイルディスクリプタをクローズする
	// ここはちゃんと返り値をチェックする必要があるっぽい
	defer syscall.Close(fd)

	// epollを使うため、Nonblockモードにする
	err = syscall.SetNonblock(fd, true)
	chk.SE(err, "Nonblockモード設定失敗")

	// address bind and listen
	addr := syscall.SockaddrInet4{
		Port: 2001,
		Addr: [4]byte{0, 0, 0, 0}, // 0.0.0.0:2000
	}

	syscall.Bind(fd, &addr)
	syscall.Listen(fd, 10)

	// epollの初期化 create
	// EpollCreateもあるが、差はないただのversion
	// TODO 引数の0の意味を調べる
	epfd, err := syscall.EpollCreate1(0)
	chk.SE(err, "EpollCreate1 Error")
	defer syscall.Close(epfd)

	event := &syscall.EpollEvent{
		Events: syscall.EPOLLIN, // read操作をなんたら
		Fd:     int32(fd),
	}

	// epoll ctrl 待つ対象の登録 (変更などもできる)
	err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, event)
	chk.SE(err, "EpollCtl Error")

	log.Println("event is ", jsonutil.Marshal(event))
	log.Println("fd is ", fd)

	userMap := map[int]*User{}

	for {
		// epoll wait 実際に待つ
		// msec: -1で待つ時間は無限に設定,タイムアウトさせない
		nEventList, err := syscall.EpollWait(epfd, eventList[:], -1)
		chk.SE(err, "EpollWait Error")

		log.Println("eventList is ", eventList)

		for ev := 0; ev < nEventList; ev++ {

			nEvent := eventList[ev]

			log.Println("request fd is ", nEvent.Fd)

			if int(nEvent.Fd) == fd { // socketのfdがroot的な? clientの接続の検知とかをこいつでできる

				// clientからの接続を許可するやつ
				connFd, sa, err := syscall.Accept(fd)
				chk.SE(err, "Accept Err")
				log.Println("sa is ", jsonutil.Marshal(sa)) // どのclientのSocket情報

				syscall.SetNonblock(fd, true)
				connEvent := &syscall.EpollEvent{
					Events: syscall.EPOLLIN | Epollet, // TODO なぜEpolletをフラグで渡すのかを検証する
					Fd:     int32(connFd),
				}
				err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, connFd, connEvent)
				chk.SE(err, "EpollCtl")

				// TODO 先にここで、名前の登録をすると面白いかもね

				// userを追加する
				userMap[connFd] = &User{
					Fd: connFd,
				}

			} else { // それ以外のfdの場合は、違う動作をするよねってはなし

				// 受け取ったメッセージをどのように送信するかのfunc
				pushMessageFunc := pushMessage

				// メッセージを受け取り
				go getMessage(int(nEvent.Fd), userMap, pushMessageFunc)
			}

		}
	}

}

func getMessage(fd int, userMap map[int]*User, pushMessage func(int, []byte, map[int]*User)) {
	defer func() {
		log.Println("fd close...", fd)

		// TODO chat roomから外す処理

		syscall.Close(fd)
	}()

	var buf [32 * 1024]byte
	for {
		// Read
		nbytes, err := syscall.Read(fd, buf[:])
		// 0byteが送られてきたら、接続終了
		if nbytes == 0 {
			log.Printf("connect close. fd:%d\n", fd)
			return
		}

		if err != nil {
			log.Printf("このconnectionはすでに終了している fd:%d\n", fd)
			return
		}

		chk.SE(err, "Read error")

		// bufは途中で値が変わるためcopyして書き込み用bufに移し替える
		copyedBuf := make([]byte, nbytes)
		copy(copyedBuf, buf[:nbytes])

		// send
		go pushMessage(fd, copyedBuf, userMap)
		log.Println("copy buf is ", string(copyedBuf))

	}
}

func pushMessage(sendFd int, messageBuf []byte, userMap map[int]*User) {

	// 自分以外だれもいない場合は何もしない
	if len(userMap) < 2 {
		return
	}

	for receiveFd := range userMap {

		// 自分の場合は、なにもしない
		if receiveFd == sendFd {
			continue
		}

		// write
		syscall.Write(receiveFd, messageBuf)
	}
}
