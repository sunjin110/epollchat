package chk

import "log"

// SE SystemErrorCheck
func SE(err error, msgs ...string) {
	if err != nil {
		if len(msgs) != 0 {
			for _, msg := range msgs {
				log.Println(msg)
			}
		}
		panic(err)
	}
}
