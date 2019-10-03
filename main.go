package main

import (
	"fmt"
)

func readDnie() {

	var cw CardWrapper
	var err error

	cw.Connect()
	defer cw.Disconnect()

	ca, data, err := cw.CardAccess()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("CardAccess", ca)
	PrettyPrint(data)

	dat_id_esp, err := cw.IDEsp()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("IDEsp", dat_id_esp)

	ci, err := cw.GetChipInfo()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Serial", ci)

	if err = cw.SelectEF(0x032F); err != nil {
		fmt.Println(err)
		return
	}

	if err = cw.GetResponse(); err != nil {
		panic(err)
	}
	cw.PrettyPrint("CERT 1f60")
}

func main() {
	readDnie()
}
